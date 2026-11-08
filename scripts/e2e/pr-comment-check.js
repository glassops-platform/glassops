// Reads the pull request comments the PR comment end-to-end run (scripts/e2e/pr-comment-e2e.sh)
// is about, through the GitHub REST API. It never prints a token or a comment body.
//
//   node pr-comment-check.js clear              delete the report comments the Actions bot wrote
//   node pr-comment-check.js snapshot <file>    expect exactly one report comment, save its id
//   node pr-comment-check.js same <file>        expect the same single comment with a new body
//
// Environment: GITHUB_TOKEN, GITHUB_REPOSITORY (owner/name), E2E_PR_NUMBER, and optionally
// GITHUB_API_URL and E2E_ACTOR (a login that must not appear in the comment). The report
// comment is the one whose body holds the marker line below. Exit 1 is a failed expectation,
// exit 2 is an API or setup error.
"use strict";

const fs = require("fs");
const crypto = require("crypto");

const digest = (text) => crypto.createHash("sha256").update(text).digest("hex");

const MARKER = "<!-- glassops-report -->";
const BOT = "github-actions[bot]";

const { GITHUB_TOKEN, GITHUB_REPOSITORY, E2E_PR_NUMBER, E2E_ACTOR } = process.env;
const API = (process.env.GITHUB_API_URL || "https://api.github.com").replace(/\/+$/, "");
for (const [name, value] of [["GITHUB_TOKEN", GITHUB_TOKEN], ["GITHUB_REPOSITORY", GITHUB_REPOSITORY], ["E2E_PR_NUMBER", E2E_PR_NUMBER]]) {
  if (!value) {
    console.error(name + " is not set");
    process.exit(2);
  }
}

async function call(method, path) {
  const res = await fetch(API + path, {
    method,
    headers: {
      Authorization: "Bearer " + GITHUB_TOKEN,
      Accept: "application/vnd.github+json",
      "X-GitHub-Api-Version": "2022-11-28",
      "User-Agent": "glassops-e2e",
    },
  });
  if (!res.ok) {
    console.error(method + " " + path + " failed: HTTP " + res.status);
    process.exit(2);
  }
  return res.status === 204 ? null : res.json();
}

// reportComments lists every comment on the pull request whose body holds the marker.
async function reportComments() {
  const found = [];
  for (let page = 1; page <= 20; page++) {
    const batch = await call("GET", "/repos/" + GITHUB_REPOSITORY + "/issues/" + E2E_PR_NUMBER + "/comments?per_page=100&page=" + page);
    for (const c of batch) {
      if (typeof c.body === "string" && c.body.includes(MARKER)) found.push(c);
    }
    if (batch.length < 100) break;
  }
  return found;
}

// leaks names what must never appear in the comment.
function leaks(body) {
  const out = [];
  if (!body.startsWith(MARKER)) out.push("the body does not start with the marker");
  if (E2E_ACTOR) {
    // The login as a whole word, so a short login does not match inside other words.
    const escaped = E2E_ACTOR.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
    if (new RegExp("(^|[^A-Za-z0-9-])" + escaped + "($|[^A-Za-z0-9-])", "i").test(body)) out.push("the body names the actor");
  }
  if (/[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}/.test(body)) out.push("the body holds an email address");
  if (/\b00D[A-Za-z0-9]{12,15}\b/.test(body)) out.push("the body holds something shaped like an org Id");
  if (body.includes(GITHUB_TOKEN)) out.push("the body holds the token");
  return out;
}

async function expectOne(what) {
  const found = await reportComments();
  if (found.length !== 1) {
    console.error("FAIL: " + what + ": want one report comment, found " + found.length);
    process.exit(1);
  }
  const problems = leaks(found[0].body);
  for (const p of problems) console.error("FAIL: " + what + ": " + p);
  if (problems.length > 0) process.exit(1);
  return found[0];
}

async function main() {
  const [cmd, file] = process.argv.slice(2);
  switch (cmd) {
    case "clear": {
      let n = 0;
      for (const c of await reportComments()) {
        if (c.user && c.user.login === BOT) {
          await call("DELETE", "/repos/" + GITHUB_REPOSITORY + "/issues/comments/" + c.id);
          n++;
        }
      }
      console.log("ok: deleted " + n + " earlier report comment(s) by " + BOT);
      break;
    }
    case "snapshot": {
      const c = await expectOne("after the first report");
      fs.writeFileSync(file, JSON.stringify({ id: c.id, body_sha256: digest(c.body) }) + "\n");
      console.log("ok: the first report created one comment");
      break;
    }
    case "same": {
      const first = JSON.parse(fs.readFileSync(file, "utf8"));
      const c = await expectOne("after the second report");
      if (c.id !== first.id) {
        console.error("FAIL: the second report made a new comment instead of updating the first");
        process.exit(1);
      }
      if (digest(c.body) === first.body_sha256) {
        console.error("FAIL: the comment did not change, so the second report did not update it");
        process.exit(1);
      }
      console.log("ok: the second report updated the same comment, and there is still one");
      break;
    }
    default:
      console.error("usage: node pr-comment-check.js clear|snapshot <file>|same <file>");
      process.exit(2);
  }
}

main().catch((err) => {
  console.error("pr-comment-check: " + err.message);
  process.exit(2);
});
