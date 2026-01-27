import { IdentityResolver } from "./identity";
import * as exec from "@actions/exec";
import * as fs from "fs";

jest.mock("@actions/exec");
jest.mock("@actions/core");
jest.mock("fs", () => ({
  ...jest.requireActual("fs"),
  writeFileSync: jest.fn(),
  existsSync: jest.fn(),
  unlinkSync: jest.fn(),
}));

describe("IdentityResolver", () => {
  let resolver: IdentityResolver;
  const mockedExec = exec.exec as jest.Mock;
  const mockedWriteFileSync = fs.writeFileSync as jest.Mock;
  const mockedExistsSync = fs.existsSync as jest.Mock;
  const mockedUnlinkSync = fs.unlinkSync as jest.Mock;

  beforeEach(() => {
    resolver = new IdentityResolver();
    jest.clearAllMocks();
  });

  describe("authenticate", () => {
    const mockAuthRequest = {
      clientId: "test-client-id",
      jwtKey:
        "-----BEGIN RSA PRIVATE KEY-----\ntest-key\n-----END RSA PRIVATE KEY-----",
      username: "test@example.com",
      instanceUrl: "https://login.salesforce.com",
    };

    it("should successfully authenticate and return org ID", async () => {
      const mockOrgId = "00D123456789012345";
      mockedExec.mockImplementation(
        async (
          cmd: string,
          args: string[],
          options: { listeners?: { stdout?: (data: Buffer) => void } },
        ) => {
          if (options?.listeners?.stdout) {
            options.listeners.stdout(
              Buffer.from(
                JSON.stringify({
                  result: { orgId: mockOrgId, accessToken: "mock-token" },
                }),
              ),
            );
          }
          return 0;
        },
      );
      mockedExistsSync.mockReturnValue(true);

      const orgId = await resolver.authenticate(mockAuthRequest);

      expect(orgId).toBe(mockOrgId);
      expect(mockedWriteFileSync).toHaveBeenCalledWith(
        expect.stringContaining("glassops-jwt"),
        mockAuthRequest.jwtKey,
        { mode: 0o600 },
      );
      expect(mockedUnlinkSync).toHaveBeenCalled();
    });

    it("should include instance URL in args when provided", async () => {
      mockedExec.mockImplementation(
        async (
          cmd: string,
          args: string[],
          options: { listeners?: { stdout?: (data: Buffer) => void } },
        ) => {
          expect(args).toContain("--instance-url");
          expect(args).toContain("https://login.salesforce.com");
          if (options?.listeners?.stdout) {
            options.listeners.stdout(
              Buffer.from(
                JSON.stringify({
                  result: { orgId: "00D123456789012345", accessToken: "mock" },
                }),
              ),
            );
          }
          return 0;
        },
      );
      mockedExistsSync.mockReturnValue(true);

      await resolver.authenticate(mockAuthRequest);

      expect(mockedExec).toHaveBeenCalledWith(
        "sf",
        expect.arrayContaining([
          "--instance-url",
          "https://login.salesforce.com",
        ]),
        expect.any(Object),
      );
    });

    it("should not include instance URL when not provided", async () => {
      const requestWithoutUrl = {
        clientId: "test-client-id",
        jwtKey: "test-key",
        username: "test@example.com",
      };

      mockedExec.mockImplementation(
        async (
          cmd: string,
          args: string[],
          options: { listeners?: { stdout?: (data: Buffer) => void } },
        ) => {
          expect(args).not.toContain("--instance-url");
          if (options?.listeners?.stdout) {
            options.listeners.stdout(
              Buffer.from(
                JSON.stringify({
                  result: { orgId: "00D123456789012345", accessToken: "mock" },
                }),
              ),
            );
          }
          return 0;
        },
      );
      mockedExistsSync.mockReturnValue(true);

      await resolver.authenticate(requestWithoutUrl);
    });

    it("should throw error on authentication failure", async () => {
      mockedExec.mockRejectedValue(new Error("Auth failed"));
      mockedExistsSync.mockReturnValue(true);

      await expect(resolver.authenticate(mockAuthRequest)).rejects.toThrow(
        "Authentication Failed",
      );
    });

    it("should clean up JWT key file even on failure", async () => {
      mockedExec.mockRejectedValue(new Error("Auth failed"));
      mockedExistsSync.mockReturnValue(true);

      try {
        await resolver.authenticate(mockAuthRequest);
      } catch {
        // Expected to throw
      }

      expect(mockedUnlinkSync).toHaveBeenCalled();
    });

    it("should not attempt to delete key file if it does not exist", async () => {
      mockedExec.mockRejectedValue(new Error("Auth failed"));
      mockedExistsSync.mockReturnValue(false);

      try {
        await resolver.authenticate(mockAuthRequest);
      } catch {
        // Expected to throw
      }

      expect(mockedUnlinkSync).not.toHaveBeenCalled();
    });
  });
});
