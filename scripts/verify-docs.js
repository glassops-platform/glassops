const fs = require('fs');
const path = require('path');
const levenshtein = require('fast-levenshtein');
const ignore = require('ignore');

const rootDir = path.resolve(__dirname, '..');
const docsGovernancePath = path.join(rootDir, 'config', 'docs-governance.json');
// Store the docs map under config instead of root
const docsStructurePath = path.join(rootDir, 'config', 'docs-structure.json');

// Default governance rules (fallback)
let governanceConfig = {
    global: {
        requiredSections: [],
        maxStalenessDays: 90,
        coverageThreshold: 0.8
    },
    system: {
        ignoreDirs: ['node_modules', '.git', '.gemini', '.glassops', 'dist', 'coverage'],
        adrPatterns: ["**/adr/*.md", "docs/adr/*.md"]
    },
    linkValidation: {
        enforceCasing: true,
        allowLocalhost: false,
        requireHttps: false,
        checkAnchors: true
    },
    rules: []
};

// State
let hasErrors = false;
let hasWarnings = false;
const docsMap = {}; // Absolute Path -> { anchors: Set<string>, links: Array<{url, line}> }
let fixPathsMode = false;
let ig = ignore();

// -------------------------------------------------------------------------
// HELPER FUNCTIONS
// -------------------------------------------------------------------------

function loadConfig() {
    if (fs.existsSync(docsGovernancePath)) {
        try {
            governanceConfig = JSON.parse(fs.readFileSync(docsGovernancePath, 'utf8'));
        } catch (e) {
            console.error('[ERROR] Failed to parse docs-governance.json:', e.message);
            process.exit(1);
        }
    }
}

function normalizeAnchor(text) {
    return text
        .trim()
        .toLowerCase()
        .replace(/[^\w\s-]/g, '') // Remove punctuation except hyphen
        .replace(/\s+/g, '-');    // Replace spaces with hyphens
}

function matchesPattern(filePath, pattern) {
    const relPath = path.relative(rootDir, filePath).replace(/\\/g, '/');
    
    let regexStr = pattern
        .replace(/\./g, '\\.')
        .replace(/\*\*/g, '___GRAVITY_DOUBLE_STAR___')
        .replace(/\*/g, '[^/]*') // Single star matches non-slash chars
        .replace(/___GRAVITY_DOUBLE_STAR___/g, '.*')
        .replace(/\!\(([^)]+)\)/g, '(?!$1)[^/]+'); // Negative lookahead requiring char match

    const regex = new RegExp(`^${regexStr}$`);
    
    // Check match against relative path (ensure leading slash for consistency if pattern implies it)
    return regex.test(relPath) || regex.test('/' + relPath);
}

function getRequiredSections(filePath) {
    for (const rule of governanceConfig.rules) {
        if (matchesPattern(filePath, rule.pattern)) {
            return rule.requiredSections;
        }
    }
    return governanceConfig.global.requiredSections;
}

// -------------------------------------------------------------------------
// CORE LOGIC
// -------------------------------------------------------------------------

async function scanDirectory(dir) {
    const entries = await fs.promises.readdir(dir, { withFileTypes: true });
    
    const tasks = entries.map(async (entry) => {
        const fullPath = path.join(dir, entry.name);
        const relPath = path.relative(rootDir, fullPath).replace(/\\/g, '/');
        
        // Use 'ignore' package to check if path is ignored
        // We explicitly check directories to prune traversal
        if (ig.ignores(relPath)) {
            return;
        }

        if (entry.isDirectory()) {
             await scanDirectory(fullPath);
        } else if (entry.isFile() && entry.name.endsWith('.md')) {
            await processFile(fullPath);
        }
    });

    await Promise.all(tasks);
}

async function processFile(filePath) {
    const content = await fs.promises.readFile(filePath, 'utf8');
    const anchors = new Set();
    const links = [];
    const relPath = path.relative(rootDir, filePath).replace(/\\/g, '/');

    const headerRegex = /^#{1,6}\s+(.+)$/gm;
    let match;
    while ((match = headerRegex.exec(content)) !== null) {
        const anchor = normalizeAnchor(match[1]);
        // Duplicate anchor check
        if (anchors.has(anchor)) {
             console.warn(`[WARN] Duplicate anchor '#${anchor}' detected in ${relPath}. Links to this section may be ambiguous.`);
             hasWarnings = true;
        }
        anchors.add(anchor);
    }

    // Regex explanation:
    // (?<!\!|\\) : Negative lookbehind to ensure [ is not preceded by ! (image) or \ (escaped)
    // \[([^\]]+)\] : Match text inside [ ]
    // \(([^)]+)\) : Match URL inside ( )
    const linkRegex = /(?<!\!|\\)\[([^\]]+)\]\(([^)]+)\)/g;

    while ((match = linkRegex.exec(content)) !== null) {
        const lineNo = content.substring(0, match.index).split('\n').length;
        // console.log(`[DEBUG] Found link in ${relPath}:${lineNo}: "${match[0]}" -> URL: "${match[2]}"`);
        links.push({
            original: match[0],
            text: match[1],
            url: match[2].trim(),
            lineNo
        });
    }

    docsMap[relPath] = {
        anchors: Array.from(anchors),
        links
    };
}

// -------------------------------------------------------------------------
// MODES
// -------------------------------------------------------------------------

async function generateMap() {
    console.log('[INFO] Scanning repository to generate baseline map...');
    await scanDirectory(rootDir);
    
    fs.writeFileSync(docsStructurePath, JSON.stringify(docsMap, null, 2));
    console.log(`[INFO] Baseline map generated at: ${docsStructurePath}`);
}

// Helper to parse .gitignore
function loadGitIgnore() {
    const gitignorePath = path.join(rootDir, '.gitignore');
    if (!fs.existsSync(gitignorePath)) return '';
    return fs.readFileSync(gitignorePath, 'utf8');
}

async function validateDocs() {
     // Prepare ignore instance
    ig = ignore();
    const systemIgnores = governanceConfig.system?.ignoreDirs || [];
    ig.add(systemIgnores);
    ig.add(loadGitIgnore());

    // Optimization: Load map if available
    if (fs.existsSync(docsStructurePath)) {
        try {
            console.log(`[INFO] Loading baseline map from ${docsStructurePath}...`);
            const cached = JSON.parse(fs.readFileSync(docsStructurePath, 'utf8'));
            Object.assign(docsMap, cached);
        } catch (e) {
            console.warn('[WARN] Cache corrupted, rescanning...');
            await scanDirectory(rootDir);
        }
    } else {
        console.log('[INFO] No baseline map found. Scanning repository...');
        await scanDirectory(rootDir);
    }

    console.log('[INFO] Validating documentation...');

    await validateAdrIndex();

    // Use async file reading
    const validationTasks = Object.entries(docsMap).map(async ([relPath, data]) => {
        const filePath = path.resolve(rootDir, relPath);
        
        // If file doesn't exist (deleted since map generation), skip
        try {
            const content = await fs.promises.readFile(filePath, 'utf8');
            
            checkGovernance(filePath, content, data.anchors);
            
            let newContent = content;
            let fileChanged = false;

            for (const link of data.links) {
                const validationResult = validateLink(filePath, link);
                
                if (validationResult.error) {
                    console.error(`[ERROR] ${relPath}:${link.lineNo} - ${validationResult.error} (Link text: "${link.text}", URL: "${link.url}")`);
                    if (validationResult.suggestion) {
                        console.log(`[INFO] Suggestion: ${validationResult.suggestion}`);
                    }
                    hasErrors = true;
                } else if (validationResult.warning) {
                    console.warn(`[WARN] ${relPath}:${link.lineNo} - ${validationResult.warning}`);
                    hasWarnings = true;
                }

                if (fixPathsMode) {
                    if (validationResult.fixedPath) {
                        const newLinkStr = `[${link.text}](${validationResult.fixedPath})`;
                        if (newContent.includes(link.original)) {
                            // Use split/join for global replacement or simple replace if unique
                            // For safety, let's use global regex replacement if possible or just replaceAll (Node 15+)
                            newContent = newContent.split(link.original).join(newLinkStr);
                            console.log(`[FIX] Updated link path in ${relPath}: ${link.url} -> ${validationResult.fixedPath}`);
                            fileChanged = true;
                        }
                    } else if (validationResult.fixedText) {
                        // Handle text fix
                        const newLinkStr = `[${validationResult.fixedText}](${link.url})`;
                         if (newContent.includes(link.original)) {
                            newContent = newContent.split(link.original).join(newLinkStr);
                            console.log(`[FIX] Updated link text in ${relPath}: ${link.text} -> ${validationResult.fixedText}`);
                            fileChanged = true;
                        }
                    }
                }
            }

            if (fixPathsMode && fileChanged) {
                await fs.promises.writeFile(filePath, newContent);
            }
        } catch (err) {
            if (err.code !== 'ENOENT') {
                console.error(`[ERROR] Failed to process ${relPath}: ${err.message}`);
            }
        }
    });

    await Promise.all(validationTasks);
}

function checkGovernance(filePath, content, anchors) {
    const relPath = path.relative(rootDir, filePath);
    const required = getRequiredSections(filePath);
    
    const normalizedAnchors = new Set(anchors); 
    const missing = [];

    for (const sec of required) {
        if (!normalizedAnchors.has(normalizeAnchor(sec))) {
            missing.push(sec);
        }
    }

    if (missing.length > 0) {
        console.error(`[ERROR] ${relPath} - Missing required sections: ${missing.join(', ')}`);
        hasErrors = true;
    }

    const stats = fs.statSync(filePath);
    const daysOld = (Date.now() - stats.mtimeMs) / (1000 * 60 * 60 * 24);
    if (daysOld > governanceConfig.global.maxStalenessDays) {
        console.warn(`[WARN] ${relPath} - Documentation may be stale (${Math.floor(daysOld)} days old)`);
        hasWarnings = true;
    }
}

function validateLink(sourceFile, link) {
    let url = link.url;
    const rules = governanceConfig.linkValidation || {};

    if (url.startsWith('http') || url.startsWith('mailto:')) {
        if (rules.requireHttps && url.startsWith('http:')) {
            return { valid: false, error: `Insecure HTTP link detected: ${url}` };
        }
        return { valid: true };
    }

    if (!rules.allowLocalhost && (url.includes('localhost') || url.includes('127.0.0.1'))) {
        return { valid: false, error: `Localhost link detected: ${url}` };
    }

    // Double hash check
    if (url.split('#').length > 2) {
         return { valid: false, warning: `Double anchor detected in link: ${url}. Only the first will be respected.` };
    }

    let linkPath = url;
    let linkAnchor = null;
    
    // Handle anchors and duplicate anchors
    if (url.includes('#')) {
        const parts = url.split('#');
        linkPath = parts[0];
        // Take the first anchor part, ignore duplicates
        if (parts.length > 1) {
            linkAnchor = normalizeAnchor(parts[1]);
        }
    }

    const targetAbsPath = !linkPath ? sourceFile : path.resolve(path.dirname(sourceFile), linkPath);
    const targetRelPath = path.relative(rootDir, targetAbsPath).replace(/\\/g, '/');
    
    // Check File Existence
    if (!docsMap[targetRelPath]) {
         const isMarkdown = targetAbsPath.toLowerCase().endsWith('.md');
         if (fs.existsSync(targetAbsPath) && !isMarkdown) {
             return { valid: true };
         }
         
         const targetBasename = path.basename(linkPath);
         let foundCandidate = null;
         
         if (targetBasename.endsWith('.md')) {
            const targetLower = targetBasename.toLowerCase();
            const candidates = [];
            
            for (const existingPath of Object.keys(docsMap)) {
                if (path.basename(existingPath).toLowerCase() === targetLower) {
                    candidates.push(existingPath);
                }
            }

            if (candidates.length === 1) {
                foundCandidate = candidates[0];
            } else if (candidates.length > 1) {
                return {
                    valid: false,
                    error: `Ambiguous link to '${targetBasename}'. Found ${candidates.length} candidates.`,
                    suggestion: `Please specify valid path to one of:\n${candidates.map(c => ' - ' + c).join('\n')}`
                };
            }
        }

        if (foundCandidate) {
             const fixedPath = path.relative(path.dirname(sourceFile), path.resolve(rootDir, foundCandidate)).replace(/\\/g, '/');
              return { 
                 valid: false, 
                 error: `Broken link: ${linkPath}`,
                 suggestion: `Found file at: ${foundCandidate}`,
                 fixedPath: fixedPath + (linkAnchor ? `#${linkAnchor}` : '')
             };
         }

        if (fs.existsSync(targetAbsPath)) {
            return { valid: true };
        }

        return { valid: false, error: `Broken link to file: ${linkPath}` };
    }

    if (linkAnchor && rules.checkAnchors !== false) {
        const targetData = docsMap[targetRelPath];
        if (!targetData.anchors.includes(linkAnchor)) {
            const bestMatch = findClosestAnchor(linkAnchor, targetData.anchors);
            
            if (bestMatch) {
                return {
                    valid: false,
                    error: `Broken anchor #${linkAnchor} in ${linkPath || 'current file'}`,
                    suggestion: `Did you mean #${bestMatch}?`
                };
            } else {
                 return {
                    valid: false,
                    error: `Broken anchor #${linkAnchor} in ${linkPath || 'current file'}. No similar anchors found.`,
                };
            }
        }
    }

    // Check canonical relative path
    if (linkPath) {
        const correctRelPath = path.relative(path.dirname(sourceFile), targetAbsPath).replace(/\\/g, '/');
        let fixedLink = correctRelPath + (linkAnchor ? `#${linkAnchor}` : '');
        if (linkPath !== correctRelPath) {
            return {
                valid: false,
                error: `Suboptimal relative path: ${linkPath}`,
                suggestion: `Should be: ${fixedLink}`,
                fixedPath: fixedLink
            };
        }

        // Configurable Link Text Check
        if (rules.enforceCasing !== false) {
            const linkText = link.text;
            // User requested: If text looks like a file (ends in .md), it should match the filename
            if (linkText.toLowerCase().endsWith('.md')) {
                 const targetBasename = path.basename(targetAbsPath);
                 if (linkText !== targetBasename) {
                     return {
                         valid: false,
                         error: `Link text '${linkText}' does not match target filename '${targetBasename}'`,
                         suggestion: `Change link text to '${targetBasename}'`,
                         // Enable auto-fix by providing the "correct" link structure (reusing targetRelPath logic if needed, but here we just need to signal the text change)
                         fixedText: targetBasename 
                     };
                 }
            }
        }
    }

    return { valid: true };
}

function findClosestAnchor(target, candidates) {
    if (!candidates || candidates.length === 0) return null;
    let closest = null;
    let minDist = Infinity;
    for (const cand of candidates) {
        const dist = levenshtein.get(target, cand);
        if (dist < minDist) {
            minDist = dist;
            closest = cand;
        }
    }
    return closest;
}

async function validateAdrIndex() {
    const adrIndexRelPath = 'docs/adr-index.md';
    if (!docsMap[adrIndexRelPath]) {
        console.error('[ERROR] adr-index.md not found in repository!');
        hasErrors = true;
        return;
    }

    const indexContent = fs.readFileSync(path.join(rootDir, adrIndexRelPath), 'utf8');
    const indexLinks = new Set();
    const linkRegex = /\[([^\]]+)\]\(([^)]+)\)/g;
    let match;
    while ((match = linkRegex.exec(indexContent)) !== null) {
        // Resolve link relative to adr-index.md (root)
        let linkUrl = match[2].trim();
        // clear anchors
        linkUrl = linkUrl.split('#')[0];
        const absPath = path.resolve(path.dirname(path.join(rootDir, adrIndexRelPath)), linkUrl);
        const relPath = path.relative(rootDir, absPath).replace(/\\/g, '/');
        indexLinks.add(relPath);
    }

    // Find all ADRs in the map using Configured Patterns
    const adrPatterns = governanceConfig.system?.adrPatterns || ["**/adr/*.md"];
    
    const potentialAdrs = Object.keys(docsMap).filter(p => {
        // Check pattern matches
        if (p.endsWith('README.md') || p.endsWith('adr-index.md')) return false;
        
        return adrPatterns.some(pattern => matchesPattern(path.join(rootDir, p), pattern));
    });

    for (const adrPath of potentialAdrs) {
        if (!indexLinks.has(adrPath)) {
            console.error(`[ERROR] ADR not found in adr-index.md: ${adrPath}`);
            hasErrors = true;
        }
    }
}


// -------------------------------------------------------------------------
// ENTRY POINT
// -------------------------------------------------------------------------

(async () => {
    const args = process.argv.slice(2);
    const command = args[0];

    loadConfig();

    if (command === 'generate') {
        // Init ignore for generate
        ig = ignore();
        const systemIgnores = governanceConfig.system?.ignoreDirs || [];
        ig.add(systemIgnores);
        ig.add(loadGitIgnore());
        
        await generateMap();
    } else if (command === 'validate' || !command) {
        if (args.includes('--fix-paths')) {
            fixPathsMode = true;
            console.log('[INFO] Auto-fix mode enabled: Will rewrite broken relative paths.');
        }
        await validateDocs();
    } else {
        console.error('[ERROR] Unknown command. Use "generate", "validate", or "validate --fix-paths"');
        process.exit(1);
    }

    console.log('\n' + '='.repeat(60));
    if (hasErrors) {
        console.error('[ERROR] Verification failed with errors.');
        process.exit(1);
    } else if (hasWarnings) {
        console.warn('[WARN] Verification passed with warnings.');
        process.exit(0);
    } else {
        console.log('[INFO] Verification passed!');
        process.exit(0);
    }
})();
