import { healthCheck } from "./health";
import * as exec from "@actions/exec";

jest.mock("@actions/exec");

describe("healthCheck", () => {
  const mockedExec = exec.exec as jest.MockedFunction<typeof exec.exec>;

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it("should return healthy with version when CLI is available", async () => {
    mockedExec.mockImplementation(async (_cmd, _args, options) => {
      const response = JSON.stringify({ cliVersion: "2.50.0" });
      options?.listeners?.stdout?.(Buffer.from(response));
      return 0;
    });

    const result = await healthCheck();

    expect(result.healthy).toBe(true);
    expect(result.version).toBe("2.50.0");
    expect(result.error).toBeUndefined();
  });

  it("should return healthy with result.cliVersion format", async () => {
    mockedExec.mockImplementation(async (_cmd, _args, options) => {
      const response = JSON.stringify({ result: { cliVersion: "2.51.0" } });
      options?.listeners?.stdout?.(Buffer.from(response));
      return 0;
    });

    const result = await healthCheck();

    expect(result.healthy).toBe(true);
    expect(result.version).toBe("2.51.0");
  });

  it("should return unhealthy when CLI execution fails", async () => {
    mockedExec.mockRejectedValue(new Error("Command not found"));

    const result = await healthCheck();

    expect(result.healthy).toBe(false);
    expect(result.error).toBe("Command not found");
    expect(result.version).toBeUndefined();
  });

  it("should handle non-Error exceptions", async () => {
    mockedExec.mockRejectedValue("string error");

    const result = await healthCheck();

    expect(result.healthy).toBe(false);
    expect(result.error).toBe("string error");
  });

  it("should return unknown version when format is unexpected", async () => {
    mockedExec.mockImplementation(async (_cmd, _args, options) => {
      const response = JSON.stringify({ someOtherField: "value" });
      options?.listeners?.stdout?.(Buffer.from(response));
      return 0;
    });

    const result = await healthCheck();

    expect(result.healthy).toBe(true);
    expect(result.version).toBe("unknown");
  });
});
