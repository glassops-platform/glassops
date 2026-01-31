/**
 * Contract Integration Tests
 *
 * These tests verify the DeploymentContractSchema works correctly
 * with various input configurations and edge cases.
 */

import {
  DeploymentContractSchema,
  DeploymentContract,
} from "../protocol/contract";

describe("Contract Integration Tests", () => {
  describe("Schema Validation", () => {
    it("should validate a complete deployment contract", () => {
      const contract = {
        schemaVersion: "1.0",
        meta: {
          adapter: "native",
          engine: "native" as const,
          timestamp: new Date().toISOString(),
          trigger: "push",
        },
        status: "Succeeded" as const,
        quality: {
          coverage: {
            actual: 85.5,
            required: 75,
            met: true,
          },
          tests: {
            total: 150,
            passed: 148,
            failed: 2,
          },
        },
        audit: {
          triggeredBy: "developer@example.com",
          orgId: "00D123456789012",
          repository: "org/repo",
          commit: "abc123def456",
        },
      };

      const result = DeploymentContractSchema.parse(contract);

      expect(result.schemaVersion).toBe("1.0");
      expect(result.status).toBe("Succeeded");
      expect(result.quality.coverage.met).toBe(true);
      expect(result.audit.orgId).toBe("00D123456789012");
    });

    it("should apply default schema version", () => {
      const contract = {
        meta: {
          adapter: "native",
          engine: "native" as const,
          timestamp: new Date().toISOString(),
          trigger: "manual",
        },
        status: "Succeeded" as const,
        quality: {
          coverage: { actual: 80, required: 80, met: true },
          tests: { total: 10, passed: 10, failed: 0 },
        },
        audit: {
          triggeredBy: "user",
          orgId: "00D000000000000",
          repository: "test/repo",
          commit: "123abc",
        },
      };

      const result = DeploymentContractSchema.parse(contract);

      expect(result.schemaVersion).toBe("1.0");
    });

    it("should support all engine types", () => {
      const engines = ["native", "hardis", "custom"] as const;

      for (const engine of engines) {
        const contract = {
          meta: {
            adapter: "test",
            engine,
            timestamp: new Date().toISOString(),
            trigger: "test",
          },
          status: "Succeeded" as const,
          quality: {
            coverage: { actual: 100, required: 80, met: true },
            tests: { total: 1, passed: 1, failed: 0 },
          },
          audit: {
            triggeredBy: "test",
            orgId: "00D",
            repository: "test/test",
            commit: "abc",
          },
        };

        const result = DeploymentContractSchema.parse(contract);
        expect(result.meta.engine).toBe(engine);
      }
    });

    it("should support all status values", () => {
      const statuses = ["Succeeded", "Failed", "Blocked"] as const;

      for (const status of statuses) {
        const contract = {
          meta: {
            adapter: "native",
            engine: "native" as const,
            timestamp: new Date().toISOString(),
            trigger: "test",
          },
          status,
          quality: {
            coverage: { actual: 80, required: 80, met: true },
            tests: { total: 1, passed: 1, failed: 0 },
          },
          audit: {
            triggeredBy: "test",
            orgId: "00D",
            repository: "test/test",
            commit: "abc",
          },
        };

        const result = DeploymentContractSchema.parse(contract);
        expect(result.status).toBe(status);
      }
    });

    it("should reject invalid engine type", () => {
      const contract = {
        meta: {
          adapter: "native",
          engine: "invalid-engine",
          timestamp: new Date().toISOString(),
          trigger: "test",
        },
        status: "Succeeded",
        quality: {
          coverage: { actual: 80, required: 80, met: true },
          tests: { total: 1, passed: 1, failed: 0 },
        },
        audit: {
          triggeredBy: "test",
          orgId: "00D",
          repository: "test/test",
          commit: "abc",
        },
      };

      expect(() => DeploymentContractSchema.parse(contract)).toThrow();
    });

    it("should reject coverage outside valid range", () => {
      const contract = {
        meta: {
          adapter: "native",
          engine: "native" as const,
          timestamp: new Date().toISOString(),
          trigger: "test",
        },
        status: "Succeeded" as const,
        quality: {
          coverage: { actual: 150, required: 80, met: true }, // Invalid: > 100
          tests: { total: 1, passed: 1, failed: 0 },
        },
        audit: {
          triggeredBy: "test",
          orgId: "00D",
          repository: "test/test",
          commit: "abc",
        },
      };

      expect(() => DeploymentContractSchema.parse(contract)).toThrow();
    });

    it("should reject negative test counts", () => {
      const contract = {
        meta: {
          adapter: "native",
          engine: "native" as const,
          timestamp: new Date().toISOString(),
          trigger: "test",
        },
        status: "Succeeded" as const,
        quality: {
          coverage: { actual: 80, required: 80, met: true },
          tests: { total: -1, passed: 1, failed: 0 }, // Invalid: negative
        },
        audit: {
          triggeredBy: "test",
          orgId: "00D",
          repository: "test/test",
          commit: "abc",
        },
      };

      expect(() => DeploymentContractSchema.parse(contract)).toThrow();
    });

    it("should reject invalid timestamp format", () => {
      const contract = {
        meta: {
          adapter: "native",
          engine: "native" as const,
          timestamp: "not-a-valid-timestamp",
          trigger: "test",
        },
        status: "Succeeded" as const,
        quality: {
          coverage: { actual: 80, required: 80, met: true },
          tests: { total: 1, passed: 1, failed: 0 },
        },
        audit: {
          triggeredBy: "test",
          orgId: "00D",
          repository: "test/test",
          commit: "abc",
        },
      };

      expect(() => DeploymentContractSchema.parse(contract)).toThrow();
    });
  });

  describe("Coverage Calculation Integration", () => {
    it("should correctly determine coverage met status", () => {
      const scenarios = [
        { actual: 85, required: 80, expected: true },
        { actual: 80, required: 80, expected: true },
        { actual: 79.9, required: 80, expected: false },
        { actual: 100, required: 100, expected: true },
        { actual: 0, required: 0, expected: true },
      ];

      for (const scenario of scenarios) {
        const contract: DeploymentContract = DeploymentContractSchema.parse({
          meta: {
            adapter: "native",
            engine: "native",
            timestamp: new Date().toISOString(),
            trigger: "test",
          },
          status: "Succeeded",
          quality: {
            coverage: {
              actual: scenario.actual,
              required: scenario.required,
              met: scenario.actual >= scenario.required,
            },
            tests: { total: 1, passed: 1, failed: 0 },
          },
          audit: {
            triggeredBy: "test",
            orgId: "00D",
            repository: "test/test",
            commit: "abc",
          },
        });

        expect(contract.quality.coverage.met).toBe(scenario.expected);
      }
    });
  });
});
