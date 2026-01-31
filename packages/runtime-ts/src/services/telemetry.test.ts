/**
 * Tests for the OpenTelemetry integration service.
 */

import { withSpan, initTelemetry, shutdown, addSpanEvent, getCurrentSpan } from "./telemetry";

// Mock the OpenTelemetry modules
jest.mock("@opentelemetry/sdk-node", () => ({
  NodeSDK: jest.fn().mockImplementation(() => ({
    start: jest.fn().mockResolvedValue(undefined),
    shutdown: jest.fn().mockResolvedValue(undefined),
  })),
}));

jest.mock("@opentelemetry/exporter-trace-otlp-http", () => ({
  OTLPTraceExporter: jest.fn(),
}));

jest.mock("@opentelemetry/resources", () => ({
  Resource: jest.fn(),
}));

describe("telemetry", () => {
  const originalEnv = process.env;

  beforeEach(() => {
    jest.clearAllMocks();
    process.env = { ...originalEnv };
  });

  afterEach(async () => {
    await shutdown();
    process.env = originalEnv;
  });

  describe("initTelemetry", () => {
    it("should not initialize when OTEL_EXPORTER_OTLP_ENDPOINT is not set", async () => {
      delete process.env.OTEL_EXPORTER_OTLP_ENDPOINT;

      await initTelemetry();

      // No SDK should be created
      const { NodeSDK } = require("@opentelemetry/sdk-node");
      expect(NodeSDK).not.toHaveBeenCalled();
    });

    it("should initialize when OTEL_EXPORTER_OTLP_ENDPOINT is set", async () => {
      process.env.OTEL_EXPORTER_OTLP_ENDPOINT = "http://localhost:4318";

      await initTelemetry("test-service", "1.0.0");

      const { NodeSDK } = require("@opentelemetry/sdk-node");
      expect(NodeSDK).toHaveBeenCalled();
    });

    it("should parse headers from OTEL_EXPORTER_OTLP_HEADERS", async () => {
      process.env.OTEL_EXPORTER_OTLP_ENDPOINT = "http://localhost:4318";
      process.env.OTEL_EXPORTER_OTLP_HEADERS =
        "Authorization=Basic 123,Custom-Header=Values";

      await initTelemetry("test-service", "1.0.0");

      const { OTLPTraceExporter } = require(
        "@opentelemetry/exporter-trace-otlp-http",
      );
      expect(OTLPTraceExporter).toHaveBeenCalledWith(
        expect.objectContaining({
          headers: {
            Authorization: "Basic 123",
            "Custom-Header": "Values",
          },
        }),
      );
    });
  });

  describe("withSpan", () => {
    it("should execute function and return result", async () => {
      const result = await withSpan("test.span", async () => {
        return "success";
      });

      expect(result).toBe("success");
    });

    it("should propagate errors from wrapped function", async () => {
      await expect(
        withSpan("test.error", async () => {
          throw new Error("test error");
        }),
      ).rejects.toThrow("test error");
    });

    it("should pass span to callback with attributes", async () => {
      let capturedSpan: unknown;

      await withSpan(
        "test.attributes",
        async (span) => {
          capturedSpan = span;
          return "done";
        },
        { custom: "value" },
      );

      expect(capturedSpan).toBeDefined();
    });
  });

  describe("addSpanEvent", () => {
    it("should not throw when no active span", () => {
      expect(() => addSpanEvent("test.event")).not.toThrow();
    });

    it("should not throw with attributes", () => {
      expect(() => addSpanEvent("test.event", { key: "value" })).not.toThrow();
    });
  });

  describe("getCurrentSpan", () => {
    it("should return undefined when no active span", () => {
      const span = getCurrentSpan();
      expect(span).toBeUndefined();
    });
  });

  describe("shutdown", () => {
    it("should not throw when called without initialization", async () => {
      await expect(shutdown()).resolves.not.toThrow();
    });

    it("should not throw when called multiple times", async () => {
      await shutdown();
      await expect(shutdown()).resolves.not.toThrow();
    });
  });
});
