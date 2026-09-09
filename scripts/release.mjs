#!/usr/bin/env node

import crypto from "node:crypto";
import { execFileSync } from "node:child_process";
import { readFileSync, writeFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { resolve } from "node:path";

const root = resolve(fileURLToPath(new URL("..", import.meta.url)));
const packagePath = resolve(root, "app/frontend/package.json");
const lockPath = resolve(root, "app/frontend/package-lock.json");
const packageHashPath = resolve(root, "app/frontend/package.json.md5");
const changelogPath = resolve(root, "CHANGELOG.md");

const args = new Map();
for (let index = 2; index < process.argv.length; index += 1) {
  const argument = process.argv[index];
  if (!argument.startsWith("--")) {
    throw new Error(`Unexpected argument: ${argument}`);
  }
  const key = argument.slice(2);
  const value = process.argv[index + 1];
  if (value && !value.startsWith("--")) {
    args.set(key, value);
    index += 1;
  } else {
    args.set(key, true);
  }
}

function git(...gitArgs) {
  return execFileSync("git", gitArgs, { cwd: root, encoding: "utf8" }).trim();
}

function releaseVersion(value) {
  const match = /^v?(\d+)\.(\d+)\.(\d+)$/.exec(value ?? "");
  if (!match) {
    throw new Error("Version must use the form vMAJOR.MINOR.PATCH.");
  }
  return `v${match[1]}.${match[2]}.${match[3]}`;
}

function latestTag() {
  try {
    return git("describe", "--tags", "--match", "v[0-9]*", "--abbrev=0", "HEAD");
  } catch {
    return "";
  }
}

function commitSubjects() {
  const tag = latestTag();
  const range = tag ? `${tag}..HEAD` : "HEAD";
  const output = git("log", range, "--no-merges", "--pretty=format:%s");
  return output ? output.split("\n").filter(Boolean) : [];
}

function sectionFor(subject) {
  const match = /^(feat|fix|perf|refactor|docs|test|ci|build|chore)(?:\(([^)]+)\))?!?:\s*(.+)$/i.exec(subject);
  if (!match) {
    return { name: "Changed", text: subject };
  }

  const sectionByType = {
    feat: "Added",
    fix: "Fixed",
    perf: "Changed",
    refactor: "Changed",
    docs: "Changed",
    test: "Changed",
    ci: "Changed",
    build: "Changed",
    chore: "Changed",
  };
  const scope = match[2] ? `**${match[2]}**: ` : "";
  return { name: sectionByType[match[1].toLowerCase()], text: `${scope}${match[3]}` };
}

function createReleaseSection(version, date, subjects) {
  const sections = new Map([
    ["Added", []],
    ["Changed", []],
    ["Fixed", []],
  ]);

  for (const subject of subjects) {
    const section = sectionFor(subject);
    sections.get(section.name).push(`- ${section.text}`);
  }

  const body = [...sections.entries()]
    .filter(([, entries]) => entries.length > 0)
    .map(([name, entries]) => `### ${name}\n${entries.join("\n")}`)
    .join("\n\n");

  return `## [${version.slice(1)}] - ${date}\n\n${body || "### Changed\n- Release preparation."}\n`;
}

function updateVersions(version) {
  const packageJson = JSON.parse(readFileSync(packagePath, "utf8"));
  const packageLock = JSON.parse(readFileSync(lockPath, "utf8"));
  const plainVersion = version.slice(1);

  packageJson.version = plainVersion;
  packageLock.version = plainVersion;
  if (packageLock.packages?.[""]) {
    packageLock.packages[""].version = plainVersion;
  }

  writeFileSync(packagePath, `${JSON.stringify(packageJson, null, 2)}\n`);
  writeFileSync(lockPath, `${JSON.stringify(packageLock, null, 2)}\n`);
  const hash = crypto.createHash("md5").update(readFileSync(packagePath)).digest("hex");
  writeFileSync(packageHashPath, `${hash}\n`);
}

function updateChangelog(version, section) {
  const changelog = readFileSync(changelogPath, "utf8");
  if (!changelog.startsWith("# Changelog\n")) {
    throw new Error("CHANGELOG.md does not start with the expected header.");
  }
  if (changelog.includes(`## [${version.slice(1)}]`)) {
    throw new Error(`CHANGELOG.md already contains version ${version}.`);
  }

  const firstRelease = changelog.indexOf("\n## [");
  if (firstRelease < 0) {
    throw new Error("Could not find the first release entry in CHANGELOG.md.");
  }
  const insertAt = firstRelease + 1;
  writeFileSync(changelogPath, `${changelog.slice(0, insertAt)}${section}\n${changelog.slice(insertAt)}`);
}

function notesFor(version) {
  const changelog = readFileSync(changelogPath, "utf8");
  const escaped = version.slice(1).replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const match = new RegExp(`## \\[${escaped}\\][\\s\\S]*?(?=\\n## \\[|$)`).exec(changelog);
  if (!match) {
    throw new Error(`Could not find release notes for ${version}.`);
  }
  return match[0].trim();
}

if (args.has("print-notes")) {
  process.stdout.write(`${notesFor(releaseVersion(args.get("version")))}\n`);
  process.exit(0);
}

const version = releaseVersion(args.get("version"));
if (git("tag", "--list", version)) {
  throw new Error(`Tag ${version} already exists.`);
}

const currentPackage = JSON.parse(readFileSync(packagePath, "utf8"));
const currentVersion = releaseVersion(currentPackage.version);
const numericVersion = (value) => value.slice(1).split(".").map(Number);
const [nextMajor, nextMinor, nextPatch] = numericVersion(version);
const [currentMajor, currentMinor, currentPatch] = numericVersion(currentVersion);
if (
  nextMajor < currentMajor ||
  (nextMajor === currentMajor && nextMinor < currentMinor) ||
  (nextMajor === currentMajor && nextMinor === currentMinor && nextPatch <= currentPatch)
) {
  throw new Error(`Version ${version} must be greater than the current package version ${currentVersion}.`);
}

const subjects = commitSubjects();
if (subjects.length === 0) {
  throw new Error("No commits found since the latest reachable version tag.");
}

const section = createReleaseSection(version, new Date().toISOString().slice(0, 10), subjects);
updateVersions(version);
updateChangelog(version, section);

const notesFile = args.get("notes-file");
if (notesFile) {
  writeFileSync(notesFile, section.trim() + "\n");
}

process.stdout.write(`Prepared release ${version} from ${subjects.length} commit(s).\n`);
