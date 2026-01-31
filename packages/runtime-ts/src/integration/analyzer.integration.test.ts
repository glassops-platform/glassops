/**
 * Analyzer Integration Tests
 *
 * Requirements:
 * - BR-003: Code Analyzer Integration
 * - Enforce opinionated usage of 'sf code-analyzer' over 'sf scanner'
 */

import * as exec from "@actions/exec";
import { Analyzer } from "../services/analyzer";

// Mock exec
jest.mock("@actions/exec");

describe("Analyzer Integration Tests", () => {
  let analyzer: Analyzer;
  const mockedExec = exec.exec as jest.MockedFunction<typeof exec.exec>;

  beforeEach(() => {
    jest.clearAllMocks();
    analyzer = new Analyzer();
  });

  it("should construct valid sf code-analyzer command", async () => {
    // Mock successful execution with clean JSON output
    mockedExec.mockImplementation(async (command, args, options) => {
      if (options?.listeners?.stdout) {
        options.listeners.stdout(Buffer.from("[]")); // No violations
      }
      return 0;
    });

    await analyzer.scan(["src"], "Complexity");

    expect(mockedExec).toHaveBeenCalledWith(
      "sf",
      expect.arrayContaining([
        "code-analyzer",
        "run",
        "--target",
        "src",
        "--ruleset",
        "Complexity",
      ]),
      expect.anything(),
    );
  });

  it("should parse violations from JSON output", async () => {
    const mockJson = JSON.stringify([
      {
        engine: "eslint",
        fileName: "src/bad.ts",
        violations: [
          {
            ruleName: "no-any",
            message: "Do not use any",
            severity: 1,
            line: 10,
          },
        ],
      },
    ]);

    mockedExec.mockImplementation(async (command, args, options) => {
      if (options?.listeners?.stdout) {
        options.listeners.stdout(Buffer.from(mockJson));
      }
      return 2; // Non-zero exit code on violation
    });

    const result = await analyzer.scan(["src"]);
    expect(result.violations).toHaveLength(1);
    expect(result.violations[0].rule).toBe("no-any");
  });

  it("should handle analyzer command failure gracefully", async () => {
    mockedExec.mockRejectedValue(new Error("Command not found: sf"));

    await expect(analyzer.scan(["src"])).rejects.toThrow("Command not found");
  });
});
