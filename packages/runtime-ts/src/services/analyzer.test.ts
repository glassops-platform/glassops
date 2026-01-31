import * as exec from "@actions/exec";
import * as core from "@actions/core";
import { Analyzer } from "./analyzer";

jest.mock("@actions/exec");
jest.mock("@actions/core");

describe("Analyzer Unit Tests", () => {
  let analyzer: Analyzer;
  const mockedExec = exec.exec as jest.MockedFunction<typeof exec.exec>;

  // Silence core outputs
  const mockedCoreError = core.error as jest.Mock;
  const mockedCoreWarning = core.warning as jest.Mock;

  beforeEach(() => {
    jest.clearAllMocks();
    analyzer = new Analyzer();
  });

  it("should handle exec failure gracefully", async () => {
    const error = new Error("Command failed");
    mockedExec.mockRejectedValue(error);

    await expect(analyzer.scan(["src"])).rejects.toThrow("Command failed");

    expect(mockedCoreError).toHaveBeenCalledWith(
      expect.stringContaining("Analyzer execution failed"),
    );
  });

  it("should handle JSON parsing errors in output", async () => {
    // We expect a warning log here, so ensure the mock is there to catch it
    mockedExec.mockImplementation(async (_command, _args, options) => {
      if (options?.listeners?.stdout) {
        options.listeners.stdout(Buffer.from("[ malformed json ]"));
      }
      return 1;
    });

    const result = await analyzer.scan(["src"]);
    expect(result.violations).toEqual([]);
    expect(result.exitCode).toBe(1);

    expect(mockedCoreWarning).toHaveBeenCalledWith(
      expect.stringContaining("Failed to parse analyzer output"),
    );
  });

  it("should handle mixed JSON output (e.g. debug logs mixed in)", async () => {
    const mixedOutput = `debug log\n[{"engine":"eslint","fileName":"test.ts","violations":[]}]\nmore logs`;

    mockedExec.mockImplementation(async (_command, _args, options) => {
      if (options?.listeners?.stdout) {
        options.listeners.stdout(Buffer.from(mixedOutput));
      }
      return 0;
    });

    const result = await analyzer.scan(["src"]);
    expect(result.exitCode).toBe(0);
    expect(result.violations).toEqual([]);
  });

  it("should return empty violations if output is empty", async () => {
    mockedExec.mockImplementation(async (_command, _args, _options) => {
      return 0;
    });

    const result = await analyzer.scan(["src"]);
    expect(result.violations).toEqual([]);
  });

  it("should include ruleset argument if provided", async () => {
    mockedExec.mockResolvedValue(0);
    await analyzer.scan(["src"], "Security");

    expect(mockedExec).toHaveBeenCalledWith(
      "sf",
      expect.arrayContaining(["--ruleset", "Security"]),
      expect.anything(),
    );
  });

  it("should capture and report violations correctly", async () => {
    const violationJson = JSON.stringify([
      {
        engine: "eslint",
        fileName: "src/bad.ts",
        violations: [
          {
            ruleName: "no-eval",
            message: "Eval is evil",
            severity: 2,
            line: 42,
          },
        ],
      },
    ]);

    mockedExec.mockImplementation(async (_command, _args, options) => {
      if (options?.listeners?.stdout) {
        options.listeners.stdout(Buffer.from(violationJson));
      }
      return 2;
    });

    const result = await analyzer.scan(["src"]);
    expect(result.violations).toHaveLength(1);
    expect(result.violations[0]).toEqual({
      rule: "no-eval",
      description: "Eval is evil",
      severity: 2,
      file: "src/bad.ts",
      line: 42,
    });
  });

  it("should capture stderr to aid debugging", async () => {
    mockedExec.mockImplementation(async (_command, _args, options) => {
      if (options?.listeners?.stderr) {
        options.listeners.stderr(Buffer.from("Doing some work..."));
      }
      return 0;
    });

    await analyzer.scan(["src"]);
  });
});
