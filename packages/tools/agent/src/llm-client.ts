import axios from 'axios';
import * as dotenv from 'dotenv';
import path from 'path';
import fs from 'fs';

const findEnv = (startDir: string): string | null => {
  let current = startDir;
  while (path.dirname(current) !== current) {
    const p = path.join(current, '.env');
    if (fs.existsSync(p)) return p;
    current = path.dirname(current);
  }
  return null;
};

const rootEnvPath = findEnv(process.cwd());

if (rootEnvPath) {
  dotenv.config({ path: rootEnvPath });
  console.log('🔍 glassops-agent: loading .env from', rootEnvPath);
} else {
  console.warn('⚠️ glassops-agent: No .env found in any parent directory.');
}

export interface ChatMessage {
  role: 'user' | 'model';
  parts: { text: string }[];
}

export class GeminiClient {
  private lastRequestTime: number = 0;
  private minInterval: number = 4000; // 4 seconds (~15 RPM) to avoid hitting 15k TPM too fast
  private apiKey: string;
  private model: string;
  private baseUrl: string;

  constructor(model: string = 'gemma-3-27b-it') {
    const key = process.env.GOOGLE_API_KEY || '';
    this.apiKey = key.replace(/^['"]|['"]$/g, '');
    this.model = model;
    this.baseUrl = `https://generativelanguage.googleapis.com/v1beta/models/${this.model}:generateContent`;

    if (!this.apiKey) {
      throw new Error('GOOGLE_API_KEY not found in .env');
    }
  }

  async generateContent(prompt: string, retryCount: number = 0): Promise<string> {
    // Rate Limiting Logic
    const now = Date.now();
    const timeSinceLast = now - this.lastRequestTime;
    if (timeSinceLast < this.minInterval) {
      const waitTime = this.minInterval - timeSinceLast;
      console.log(`⏳ Rate limiting: Waiting ${Math.ceil(waitTime / 1000)}s...`);
      await new Promise(resolve => setTimeout(resolve, waitTime));
    }
    
    this.lastRequestTime = Date.now();

    const payload = {
      contents: [{ parts: [{ text: prompt }] }],
      generationConfig: {
        maxOutputTokens: 8192,
        temperature: 0.2
      }
    };

    try {
      const response = await axios.post(this.baseUrl, payload, {
        headers: {
          'Content-Type': 'application/json',
          'x-goog-api-key': this.apiKey
        }
      });

      const candidate = response.data?.candidates?.[0];
      const text = candidate?.content?.parts?.[0]?.text;
      
      if (!text) {
        console.error('🚫 Gemini returned no text. Finish Reason:', candidate?.finishReason);
        console.debug('Full Response:', JSON.stringify(response.data, null, 2));
        throw new Error(`Invalid response format from Gemini (FinishReason: ${candidate?.finishReason})`);
      }
      return text;
    } catch (error: unknown) {
      if (axios.isAxiosError(error)) {
        const status = error.response?.status;
        const data = error.response?.data;

        if ((status === 429 || status === 503) && retryCount < 3) {
          const backoffs = [10000, 30000, 60000]; // 10s, 30s, 60s
          const wait = backoffs[retryCount] || 60000;
          console.warn(`⚠️ ${status} Error (Overloaded/Quota). Retrying in ${wait / 1000}s... (Attempt ${retryCount + 1}/3)`);
          await new Promise(resolve => setTimeout(resolve, wait));
          return this.generateContent(prompt, retryCount + 1);
        }

        console.error('📡 Gemini API Error:', status, data);
      } else {
        const err = error as Error;
        console.error('❌ Request error:', err.message);
      }
      throw error;
    }
  }
}
