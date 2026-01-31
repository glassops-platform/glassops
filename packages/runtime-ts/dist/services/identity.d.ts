interface AuthRequest {
    clientId: string;
    jwtKey: string;
    username: string;
    instanceUrl?: string;
}
export declare class IdentityResolver {
    authenticate(req: AuthRequest): Promise<string>;
}
export {};
//# sourceMappingURL=identity.d.ts.map