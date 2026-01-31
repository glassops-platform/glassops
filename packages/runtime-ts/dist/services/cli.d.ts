import { ProtocolConfig } from "../protocol/policy";
export declare class RuntimeEnvironment {
    protected platform: NodeJS.Platform;
    install(version?: string): Promise<void>;
    installPlugins(config: ProtocolConfig, plugins: string[]): Promise<void>;
    private execWithAutoConfirm;
}
//# sourceMappingURL=cli.d.ts.map