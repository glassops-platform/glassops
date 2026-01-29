import { DeploymentContractSchema } from "./contract";

describe("DeploymentContractSchema", () => {
  describe("schema validation", () => {
    it("should validate a complete deployment contract", () => {
      const validContract = {
        schemaVersion: "1.0",
        meta: {
          adapter: "native",
          engine: "native" as const,
          timestamp: "2024-01-21T18:00:00.000Z",
          trigger: "push",
        },
        status: "Succeeded" as const,
        quality: {
          coverage: {
            actual: 85,
            required: 80,
            met: true,
          },
          tests: {
            total: 100,
            passed: 95,
            failed: 5,
          },
        },
        audit: {
          triggeredBy: "github-user",
          orgId: "00D123456789012345",
          repository: "org/repo",
          commit: "abc123def456",
        },
      };

      const result = DeploymentContractSchema.parse(validContract);
      expect(result).toEqual(validContract);
    });

    it("should use default schemaVersion when not provided", () => {
      const contractWithoutVersion = {
        meta: {
          adapter: "native",
          engine: "native" as const,
          timestamp: "2024-01-21T18:00:00.000Z",
          trigger: "push",
        },
        status: "Succeeded" as const,
        quality: {
          coverage: {
            actual: 85,
            required: 80,
            met: true,
          },
          tests: {
            total: 100,
            passed: 95,
            failed: 5,
          },
        },
        audit: {
          triggeredBy: "github-user",
          orgId: "00D123456789012345",
          repository: "org/repo",
          commit: "abc123def456",
        },
      };

      const result = DeploymentContractSchema.parse(contractWithoutVersion);
      expect(result.schemaVersion).toBe("1.0");
    });

    it("should reject invalid engine types", () => {
      const invalidContract = {
        schemaVersion: "1.0",
        meta: {
          adapter: "native",
          engine: "invalid" as unknown, // This should fail
          timestamp: "2024-01-21T18:00:00.000Z",
          trigger: "push",
        },
        status: "Succeeded" as const,
        quality: {
          coverage: {
            actual: 85,
            required: 80,
            met: true,
          },
          tests: {
            total: 100,
            passed: 95,
            failed: 5,
          },
        },
        audit: {
          triggeredBy: "github-user",
          orgId: "00D123456789012345",
          repository: "org/repo",
          commit: "abc123def456",
        },
      };

      expect(() => DeploymentContractSchema.parse(invalidContract)).toThrow();
    });

    it("should reject invalid status values", () => {
      const invalidContract = {
        schemaVersion: "1.0",
        meta: {
          adapter: "native",
          engine: "native" as const,
          timestamp: "2024-01-21T18:00:00.000Z",
          trigger: "push",
        },
        status: "InvalidStatus" as unknown, // This should fail
        quality: {
          coverage: {
            actual: 85,
            required: 80,
            met: true,
          },
          tests: {
            total: 100,
            passed: 95,
            failed: 5,
          },
        },
        audit: {
          triggeredBy: "github-user",
          orgId: "00D123456789012345",
          repository: "org/repo",
          commit: "abc123def456",
        },
      };

      expect(() => DeploymentContractSchema.parse(invalidContract)).toThrow();
    });

    it("should reject invalid timestamp format", () => {
      const invalidContract = {
        schemaVersion: "1.0",
        meta: {
          adapter: "native",
          engine: "native" as const,
          timestamp: "2024-01-21 18:00:00", // Invalid format
          trigger: "push",
        },
        status: "Succeeded" as const,
        quality: {
          coverage: {
            actual: 85,
            required: 80,
            met: true,
          },
          tests: {
            total: 100,
            passed: 95,
            failed: 5,
          },
        },
        audit: {
          triggeredBy: "github-user",
          orgId: "00D123456789012345",
          repository: "org/repo",
          commit: "abc123def456",
        },
      };

      expect(() => DeploymentContractSchema.parse(invalidContract)).toThrow();
    });

    it("should reject negative test counts", () => {
      const invalidContract = {
        schemaVersion: "1.0",
        meta: {
          adapter: "native",
          engine: "native" as const,
          timestamp: "2024-01-21T18:00:00.000Z",
          trigger: "push",
        },
        status: "Succeeded" as const,
        quality: {
          coverage: {
            actual: 85,
            required: 80,
            met: true,
          },
          tests: {
            total: -1, // Invalid
            passed: 95,
            failed: 5,
          },
        },
        audit: {
          triggeredBy: "github-user",
          orgId: "00D123456789012345",
          repository: "org/repo",
          commit: "abc123def456",
        },
      };

      expect(() => DeploymentContractSchema.parse(invalidContract)).toThrow();
    });

    it("should reject coverage percentages outside 0-100 range", () => {
      const invalidContract = {
        schemaVersion: "1.0",
        meta: {
          adapter: "native",
          engine: "native" as const,
          timestamp: "2024-01-21T18:00:00.000Z",
          trigger: "push",
        },
        status: "Succeeded" as const,
        quality: {
          coverage: {
            actual: 150, // Invalid
            required: 80,
            met: true,
          },
          tests: {
            total: 100,
            passed: 95,
            failed: 5,
          },
        },
        audit: {
          triggeredBy: "github-user",
          orgId: "00D123456789012345",
          repository: "org/repo",
          commit: "abc123def456",
        },
      };

      expect(() => DeploymentContractSchema.parse(invalidContract)).toThrow();
    });
  });
});
