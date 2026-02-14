
import xml.etree.ElementTree as ET
from pathlib import Path
from typing import List
from .base import BaseAdapter


class XMLAdapter(BaseAdapter):
    """
    Adapter for processing XML files.
    Detects the purpose of the XML from its root element, namespace,
    and structure — then provides that context to the LLM.
    """

    # Map root elements to human-readable descriptions
    KNOWN_ROOT_ELEMENTS = {
        # Salesforce metadata
        "Flow": "Salesforce Flow (automation workflow)",
        "Profile": "Salesforce Profile (security and access control)",
        "PermissionSet": "Salesforce Permission Set (granular access grants)",
        "PermissionSetGroup": "Salesforce Permission Set Group (bundled access)",
        "CustomApplication": "Salesforce Custom Application (app shell and navigation)",
        "CustomObject": "Salesforce Custom Object (data model definition)",
        "CustomField": "Salesforce Custom Field (field definition)",
        "Layout": "Salesforce Page Layout (record page field arrangement)",
        "OmniScript": "OmniStudio OmniScript (multi-step guided process)",
        "OmniUiCard": "OmniStudio FlexCard (contextual UI card)",
        "OmniDataTransform": "OmniStudio DataRaptor (data extraction/transformation)",
        "OmniIntegrationProcedure": "OmniStudio Integration Procedure (server-side orchestration)",
        "Package": "Salesforce Package Manifest (deployment descriptor)",
        "ApexClass": "Salesforce Apex Class (metadata companion)",
        "ApexTrigger": "Salesforce Apex Trigger (metadata companion)",
        "LightningComponentBundle": "Lightning Web Component (metadata)",
        "AuraDefinitionBundle": "Aura Component (metadata)",
        "ApexPage": "Visualforce Page (metadata)",
        # Build / CI
        "project": "Project configuration",
        "configuration": "Application configuration",
        "settings": "Settings configuration",
        # Common XML formats
        "rss": "RSS feed",
        "feed": "Atom feed",
        "svg": "SVG vector graphic",
        "html": "XHTML document",
        "urlset": "XML sitemap",
    }

    TARGET_CHUNK_SIZE = 12000

    def can_handle(self, file_path: Path) -> bool:
        # Exclude metadata files that are bundled with other adapters
        if file_path.name.endswith("-meta.xml"):
            excluded_suffixes = {
                ".cls-meta.xml", ".trigger-meta.xml",  # Apex
                ".js-meta.xml",                        # LWC
                ".cmp-meta.xml",                       # Aura
                ".evt-meta.xml", ".intf-meta.xml",
                ".tokens-meta.xml"
            }
            if any(file_path.name.endswith(suffix) for suffix in excluded_suffixes):
                return False

        return file_path.suffix.lower() == ".xml"

    def _detect_context(self, file_path: Path, content: str) -> str:
        """
        Inspect the XML to determine what kind of document it is.
        Returns a context block that helps the LLM understand the purpose.
        """
        context_parts = [f"File: {file_path.name}"]

        try:
            root = ET.fromstring(content)

            # Extract root tag (strip namespace)
            raw_tag = root.tag
            if "}" in raw_tag:
                namespace = raw_tag.split("}")[0].strip("{")
                tag = raw_tag.split("}")[1]
                context_parts.append(f"Namespace: {namespace}")
            else:
                tag = raw_tag
                namespace = None

            # Look up known type or describe what we see
            if tag in self.KNOWN_ROOT_ELEMENTS:
                context_parts.insert(0, f"Detected type: {self.KNOWN_ROOT_ELEMENTS[tag]}")
            else:
                context_parts.insert(0, f"Root element: <{tag}>")

            # Structural overview — what child elements are present
            children = list(root)
            if children:
                child_tags = []
                for child in children:
                    child_tag = child.tag.split("}")[-1] if "}" in child.tag else child.tag
                    child_tags.append(child_tag)

                # Count unique child types
                unique_tags = {}
                for t in child_tags:
                    unique_tags[t] = unique_tags.get(t, 0) + 1

                # Build a summary like "3 actionCalls, 2 decisions, 1 start"
                summary_parts = []
                for t, count in sorted(unique_tags.items(), key=lambda x: -x[1])[:12]:
                    summary_parts.append(f"{count} {t}" if count > 1 else t)

                context_parts.append(
                    f"Structure: {len(children)} child elements "
                    f"({', '.join(summary_parts)})"
                )

            # Check for attributes on root that give context
            if root.attrib:
                interesting_attrs = {
                    k: v for k, v in root.attrib.items()
                    if not k.startswith("{") and k.lower() not in ("xmlns",)
                }
                if interesting_attrs:
                    attr_str = ", ".join(f'{k}="{v}"' for k, v in list(interesting_attrs.items())[:5])
                    context_parts.append(f"Root attributes: {attr_str}")

        except ET.ParseError:
            context_parts.append("(XML parsing failed — raw content provided)")

        return "\n".join(context_parts)

    def parse(self, file_path: Path, content: str) -> List[str]:
        # Detect what this XML is
        context = self._detect_context(file_path, content)
        context_header = f"<!-- CONTEXT:\n{context}\n-->\n\n"

        # If content is small enough, return as single chunk with context
        if len(content) <= self.TARGET_CHUNK_SIZE:
            return [context_header + content]

        # For large files, put context in first chunk only
        chunks = []
        current_chunk = ""

        for line in content.split("\n"):
            if len(current_chunk) + len(line) > self.TARGET_CHUNK_SIZE:
                if current_chunk.strip():
                    chunks.append(current_chunk)
                current_chunk = line + "\n"
            else:
                current_chunk += line + "\n"

        if current_chunk.strip():
            chunks.append(current_chunk)

        # Prepend context to first chunk
        if chunks:
            chunks[0] = context_header + chunks[0]

        return chunks

    def get_prompt(self, file_path: Path, parsed_content: str) -> str:
        return (
            "You are a technical documentation expert. Analyze the XML content below. "
            "A CONTEXT block at the top identifies the type of XML document. "
            "Focus on the PURPOSE and BEHAVIOR — what this configuration does, "
            "not just its XML structure. Explain it so both technical and "
            "non-technical readers understand.\n\n"
            f"{parsed_content}"
        )

    def validate_content(self, content: str) -> List[str]:
        errors = []
        try:
            ET.fromstring(content)
        except ET.ParseError as e:
            errors.append(f"XML parse error: {e}")
        return errors
