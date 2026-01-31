/**
 * Identity Integration Tests
 *
 * These tests verify the IdentityResolver class works correctly
 * with real Salesforce CLI authentication flows.
 *
 * RTM Coverage:
 * - BR-011: JWT Authentication Flow (TC-011-P01, TC-011-N01, TC-011-N02)
 */

import * as fs from "fs";
import * as os from "os";
import { IdentityResolver } from "../services/identity";
import * as exec from "@actions/exec";

// Mock external GitHub Actions modules
jest.mock("@actions/exec");
jest.mock("@actions/core");

describe("Identity Integration Tests", () => {
  let identity: IdentityResolver;
  const mockedExec = exec.exec as jest.Mock;

  beforeEach(() => {
    identity = new IdentityResolver();
    jest.clearAllMocks();

    // Mock the exec function to simulate successful authentication
    mockedExec.mockImplementation(
      async (
        command: string,
        args: string[],
        options: {
          listeners?: { stdout?: (data: Buffer) => void };
          silent?: boolean;
        },
      ) => {
        // Simulate stdout listener capturing JSON response
        if (options?.listeners?.stdout) {
          const mockResponse = JSON.stringify({
            result: { orgId: "00D123456789012345", accessToken: "test-token" },
          });
          options.listeners.stdout(Buffer.from(mockResponse));
        }
        return 0; // Success exit code
      },
    );
  });

  describe("Successful Authentication", () => {
    it("should authenticate successfully with valid credentials", async () => {
      const result = await identity.authenticate({
        clientId: "test-client-id",
        jwtKey:
          "-----BEGIN RSA PRIVATE KEY-----\ntest-key\n-----END RSA PRIVATE KEY-----",
        username: "test@example.com",
        instanceUrl: "https://login.salesforce.com",
      });

      expect(result).toBe("00D123456789012345");
      expect(mockedExec).toHaveBeenCalledWith(
        "sf",
        expect.arrayContaining([
          "org",
          "login",
          "jwt",
          "--client-id",
          "test-client-id",
          "--username",
          "test@example.com",
          "--set-default",
          "--instance-url",
          "https://login.salesforce.com",
          "--json",
        ]),
        expect.any(Object),
      );
    });

    it("should authenticate without instance URL", async () => {
      const result = await identity.authenticate({
        clientId: "test-client-id",
        jwtKey:
          "-----BEGIN RSA PRIVATE KEY-----\ntest-key\n-----END RSA PRIVATE KEY-----",
        username: "test@example.com",
      });

      expect(result).toBe("00D123456789012345");
      // Verify instance-url is NOT in the args when not provided
      const execCall = mockedExec.mock.calls[0];
      expect(execCall[1]).not.toContain("--instance-url");
    });

    it("should clean up JWT key file after successful authentication", async () => {
      const tempDir = os.tmpdir();
      const keyFilesBefore = fs
        .readdirSync(tempDir)
        .filter((f) => f.startsWith("glassops-jwt-"));

      await identity.authenticate({
        clientId: "test-client-id",
        jwtKey:
          "-----BEGIN RSA PRIVATE KEY-----\ntest-key\n-----END RSA PRIVATE KEY-----",
        username: "test@example.com",
      });

      const keyFilesAfter = fs
        .readdirSync(tempDir)
        .filter((f) => f.startsWith("glassops-jwt-"));
      expect(keyFilesAfter.length).toBe(keyFilesBefore.length);
    });
  });

  describe("Authentication Failure Scenarios", () => {
    it("should throw error when CLI command fails", async () => {
      mockedExec.mockRejectedValue(new Error("Authentication failed"));

      await expect(
        identity.authenticate({
          clientId: "test-client-id",
          jwtKey:
            "-----BEGIN RSA PRIVATE KEY-----\ntest-key\n-----END RSA PRIVATE KEY-----",
          username: "test@example.com",
        }),
      ).rejects.toThrow(
        "❌ Authentication Failed. Check Client ID and JWT Key.",
      );
    });

    it("should throw error when JSON parsing fails", async () => {
      mockedExec.mockImplementation(
        async (
          command: string,
          args: string[],
          options: {
            listeners?: { stdout?: (data: Buffer) => void };
            silent?: boolean;
          },
        ) => {
          if (options?.listeners?.stdout) {
            options.listeners.stdout(Buffer.from("invalid-json"));
          }
          return 0;
        },
      );

      await expect(
        identity.authenticate({
          clientId: "test-client-id",
          jwtKey:
            "-----BEGIN RSA PRIVATE KEY-----\ntest-key\n-----END RSA PRIVATE KEY-----",
          username: "test@example.com",
        }),
      ).rejects.toThrow(
        "❌ Authentication Failed. Check Client ID and JWT Key.",
      );
    });

    it("should clean up JWT key file even on failure", async () => {
      mockedExec.mockRejectedValue(new Error("Authentication failed"));

      const tempDir = os.tmpdir();

      try {
        await identity.authenticate({
          clientId: "test-client-id",
          jwtKey:
            "-----BEGIN RSA PRIVATE KEY-----\ntest-key\n-----END RSA PRIVATE KEY-----",
          username: "test@example.com",
        });
      } catch {
        // Expected to fail
      }

      // Verify no leftover key files
      const keyFilesAfter = fs
        .readdirSync(tempDir)
        .filter((f) => f.startsWith("glassops-jwt-"));
      expect(keyFilesAfter.length).toBe(0);
    });
  });

  describe("Different Salesforce Environments", () => {
    it("should authenticate with sandbox instance", async () => {
      const result = await identity.authenticate({
        clientId: "test-client-id",
        jwtKey:
          "-----BEGIN RSA PRIVATE KEY-----\ntest-key\n-----END RSA PRIVATE KEY-----",
        username: "test@example.com",
        instanceUrl: "https://test.salesforce.com",
      });

      expect(result).toBe("00D123456789012345");
      expect(mockedExec).toHaveBeenCalledWith(
        "sf",
        expect.arrayContaining([
          "--instance-url",
          "https://test.salesforce.com",
        ]),
        expect.any(Object),
      );
    });

    it("should authenticate with custom domain", async () => {
      const result = await identity.authenticate({
        clientId: "test-client-id",
        jwtKey:
          "-----BEGIN RSA PRIVATE KEY-----\ntest-key\n-----END RSA PRIVATE KEY-----",
        username: "test@example.com",
        instanceUrl: "https://mycompany.my.salesforce.com",
      });

      expect(result).toBe("00D123456789012345");
      expect(mockedExec).toHaveBeenCalledWith(
        "sf",
        expect.arrayContaining([
          "--instance-url",
          "https://mycompany.my.salesforce.com",
        ]),
        expect.any(Object),
      );
    });
  });

  describe("Silent Execution", () => {
    it("should execute CLI commands silently", async () => {
      await identity.authenticate({
        clientId: "test-client-id",
        jwtKey:
          "-----BEGIN RSA PRIVATE KEY-----\ntest-key\n-----END RSA PRIVATE KEY-----",
        username: "test@example.com",
      });

      const execCall = mockedExec.mock.calls[0];
      const options = execCall[2];
      expect(options.silent).toBe(true);
    });

    it("should use listeners for stdout capture", async () => {
      await identity.authenticate({
        clientId: "test-client-id",
        jwtKey:
          "-----BEGIN RSA PRIVATE KEY-----\ntest-key\n-----END RSA PRIVATE KEY-----",
        username: "test@example.com",
      });

      const execCall = mockedExec.mock.calls[0];
      const options = execCall[2];
      expect(options.listeners).toBeDefined();
      expect(options.listeners.stdout).toBeDefined();
    });
  });
});
