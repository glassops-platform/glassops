export interface AgentAdapter {
  /**
   * Identifies if this adapter can handle the given file extension.
   */
  canHandle(fileName: string): boolean;

  /**
   * Parses the file content and returns a structured representation or prompt context.
   * returning an array allows for chunking large files.
   */
  parse(filePath: string, content: string): Promise<string[]>;

  /**
   * Generates the final documentation or metadata based on the LLM's output.
   */
  generatePrompt(filePath: string, parsedContent: string): string;

  /**
   * Post-processes the LLM output (e.g., formatting, relative link fixing).
   * If the file was chunked, this method handles joining them.
   */
  postProcess(filePath: string, outputs: string[]): string;
}
