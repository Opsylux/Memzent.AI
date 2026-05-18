"use client"; // Required for interactivity in Next.js App Router

import { useState } from "react";
import { CodeBlock } from "@/components/docs/code-block";
import { ArrowRight, Key, Zap, CheckCircle2 } from "lucide-react";
import Link from "next/link";
import { DocsPager } from "@/components/docs/docs-pager";
import { DOCS_CONFIG } from "@/config/docs-config";

export default function QuickStart() {
  const curlExample = `curl -X POST https://${DOCS_CONFIG.domain}/v1/chat \\
  -H "X-API-Key: memzent_f7c9...8e2a" \\
  -H "Content-Type: application/json" \\
  -d '{
    "messages": [
      {
        "role": "user",
        "content": "Find all high-priority tickets from the last 24 hours"
      }
    ]
  }'`;

  const responseExample = `{
  "text": "There are 3 high-priority tickets opened in the last 24 hours...",
  "cached": false,
  "provider": "ollama",
  "request_id": "c92f1b0a8d437e618956e18f293b4a5d"
}`;

  const pythonExample = `import requests

url = "https://${DOCS_CONFIG.domain}/v1/chat"
headers = {
    "X-API-Key": "MEMZENT_TOKEN_KEY",
    "Content-Type": "application/json"
}
payload = {
    "messages": [{"role": "user", "content": "Explain role-based access control"}],
    "skip_cache": False
}

response = requests.post(url, json=payload, headers=headers)
print("Response:", response.json()["text"])`;

  const golangExample = `package main

import (
  "bytes"
  "fmt"
  "net/http"
  "io"
)

func main() {
  url := "https://${DOCS_CONFIG.domain}/v1/chat"
  payload := []byte(\`{"messages":[{"role":"user","content":"Explain role-based access control"}]}\`)

  req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
  req.Header.Set("X-API-Key", "MEMZENT_TOKEN_KEY")
  req.Header.Set("Content-Type", "application/json")

  client := &http.Client{}
  resp, _ := client.Do(req)
  defer resp.Body.Close()

  body, _ := io.ReadAll(resp.Body)
  fmt.Println(string(body))
}`;

  const javaExample = `import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;

public class Main {
    public static void main(String[] args) throws Exception {
        HttpClient client = HttpClient.newHttpClient();
        HttpRequest request = HttpRequest.newBuilder()
            .uri(URI.create("https://${DOCS_CONFIG.domain}/v1/chat"))
            .header("X-API-Key", "MEMZENT_TOKEN_KEY")
            .header("Content-Type", "application/json")
            .POST(HttpRequest.BodyPublishers.ofString("{\\"messages\\":[{\\"role\\":\\"user\\",\\"content\\":\\"Explain role-based access control\\"}]}"))
            .build();

        HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());
        System.out.println(response.body());
    }
}`;

  const nextjsExample = `// app/actions.ts
"use server"

export async function executePrompt(prompt: string) {
    const res = await fetch("https://${DOCS_CONFIG.domain}/v1/chat", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "X-API-Key": process.env.MEMZENT_API_KEY!
        },
        body: JSON.stringify({
            messages: [{ role: "user", content: prompt }]
        })
    });
    return res.json();
}`;
  const javascriptExample = `import axios from "axios";

const url = "https://${DOCS_CONFIG.domain}/v1/chat";
const headers = {
    "X-API-Key": "MEMZENT_TOKEN_KEY",
    "Content-Type": "application/json"
};
const payload = {
    messages: [{ role: "user", content: "Explain role-based access control" }],
    skip_cache: false
};

async function sendRequest() {
    try {
        const response = await axios.post(url, payload, { headers });
        console.log("Response:", response.data.text);
    } catch (error) {
        console.error("Error connecting to Memzent:", error.message);
    }
}

sendRequest();`;
  const dotnetExample = `using System;
using System.Net.Http;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;

class Program
{
    static async Task Main()
    {
        var url = "https://${DOCS_CONFIG.domain}/v1/chat";
        using var client = new HttpClient();
        
        client.DefaultRequestHeaders.Add("X-API-Key", "MEMZENT_TOKEN_KEY");
        
        var payload = new
        {
            messages = new[] { new { role = "user", content = "Explain role-based access control" } }
        };

        var json = JsonSerializer.Serialize(payload);
        var content = new StringContent(json, Encoding.UTF8, "application/json");

        try
        {
            var response = await client.PostAsync(url, content);
            var responseString = await response.Content.ReadAsStringAsync();
            Console.WriteLine(responseString);
        }
        catch (Exception ex)
        {
            Console.WriteLine($"Error: {ex.Message}");
        }
    }
}`;

  // Map environments to their tab properties
  const languageTabs = [
    { id: "python", label: "Python", language: "python", code: pythonExample },
    { id: ".net", label: ".NET", language: "csharp", code: dotnetExample },
    { id: "go", label: "Golang", language: "go", code: golangExample },
    { id: "java", label: "Java", language: "java", code: javaExample },
    { id: "nextjs", label: "Next.js", language: "typescript", code: nextjsExample },
    { id: "javascript", label: "JavaScript", language: "javascript", code: javascriptExample },
  ];

  const [activeTabId, setActiveTabId] = useState(languageTabs[0].id);
  const activeTab = languageTabs.find((tab) => tab.id === activeTabId) || languageTabs[0];

  return (
    <div className="space-y-12">
      <header className="space-y-4">
        <div className="flex items-center gap-2 px-3 py-1 rounded-full bg-memzent-glow/5 border border-memzent-glow/20 w-fit">
          <span className="text-[10px] font-black text-memzent-glow uppercase tracking-tighter italic">Getting_Started</span>
        </div>
        <h1 className="text-4xl font-black tracking-tighter uppercase sm:text-5xl">Quick Start Guide</h1>
        <p className="text-lg text-white/60 leading-relaxed font-medium">
          Get Memzent up and running in under 5 minutes. All you need is an API key and one HTTP request.
        </p>
      </header>

      {/* Step 1 */}
      <section className="space-y-6">
        <div className="flex items-center gap-4">
          <div className="w-8 h-8 rounded-full bg-memzent-glow/10 border border-memzent-glow/20 flex items-center justify-center text-xs font-black text-memzent-glow">1</div>
          <h2 className="text-2xl font-black tracking-tighter uppercase">Get Your API Key</h2>
        </div>
        <div className="space-y-4 pl-12">
          <p className="text-sm text-white/60 leading-relaxed font-medium">
            Go to the <a href="/keys" className="text-memzent-glow underline font-bold">API Keys</a> section of your Dashboard and click <strong className="text-white">+ Generate Secret Key</strong>. Copy it immediately — it is only shown once.
          </p>
          <div className="p-4 rounded-xl bg-memzent-glow/5 border border-memzent-glow/10 flex items-start gap-3">
            <Key size={16} className="text-memzent-glow mt-0.5 shrink-0" />
            <p className="text-xs text-memzent-glow font-bold leading-relaxed">
              Keep your API key secret. Never include it in client-side JavaScript or expose it in a public repository.
            </p>
          </div>
        </div>
      </section>

      {/* Step 2 */}
      <section className="space-y-6">
        <div className="flex items-center gap-4">
          <div className="w-8 h-8 rounded-full bg-memzent-glow/10 border border-memzent-glow/20 flex items-center justify-center text-xs font-black text-memzent-glow">2</div>
          <h2 className="text-2xl font-black tracking-tighter uppercase">Send Your First Request</h2>
        </div>
        <div className="space-y-5 pl-12">
          <p className="text-sm text-white/60 leading-relaxed font-medium">
            Send a <code className="text-memzent-glow bg-memzent-glow/5 px-1 rounded font-mono">POST</code> request to the chat endpoint. Include your key in the <code className="text-memzent-glow bg-memzent-glow/5 px-1 rounded font-mono">X-API-Key</code> header.
          </p>
          <CodeBlock code={curlExample} language="bash" filename="cURL" />

          <p className="text-sm text-white/60 leading-relaxed font-medium">You will receive a structured JSON response:</p>
          <CodeBlock code={responseExample} language="json" filename="Response" />
        </div>
      </section>

      {/* Step 3 — Code Examples */}
      <section className="space-y-6">
        <div className="flex items-center gap-4">
          <div className="w-8 h-8 rounded-full bg-memzent-glow/10 border border-memzent-glow/20 flex items-center justify-center text-xs font-black text-memzent-glow">3</div>
          <h2 className="text-2xl font-black tracking-tighter uppercase">Integrate via API</h2>
        </div>
        <div className="space-y-6 pl-12">
          <p className="text-sm text-white/60 leading-relaxed font-medium">
            Since Memzent acts as an HTTP proxy, you can connect to it using any language that supports standard HTTP requests. Here are examples for popular environments:
          </p>

          <div className="space-y-4">
            {/* Tab Headers */}
            <div className="flex flex-wrap gap-2 border-b border-white/10 pb-px">
              {languageTabs.map((tab) => {
                const isActive = activeTabId === tab.id;
                return (
                  <button
                    key={tab.id}
                    onClick={() => setActiveTabId(tab.id)}
                    type="button"
                    className={`px-4 py-2.5 text-xs font-bold uppercase tracking-wider transition-all duration-200 border-b-2 -mb-px relative ${isActive
                      ? "text-memzent-glow border-memzent-glow"
                      : "text-white/40 border-transparent hover:text-white/80"
                      }`}
                  >
                    {tab.label}
                  </button>
                );
              })}
            </div>

            {/* Active Code Panel */}
            <div className="w-full">
              <CodeBlock
                key={activeTab.id} // Forces clean block rendering when toggled
                code={activeTab.code}
                language={activeTab.language}
                filename={activeTab.label}
              />
            </div>
          </div>
        </div>
      </section>

      {/* Step 4 — Check the trace */}
      <section className="space-y-6">
        <div className="flex items-center gap-4">
          <div className="w-8 h-8 rounded-full bg-memzent-glow/10 border border-memzent-glow/20 flex items-center justify-center text-xs font-black text-memzent-glow">4</div>
          <h2 className="text-2xl font-black tracking-tighter uppercase">Inspect the Execution Trace</h2>
        </div>
        <div className="space-y-4 pl-12">
          <p className="text-sm text-white/60 leading-relaxed font-medium">
            After your first request, check your Dashboard. You will see a real-time trace showing how Memzent processed the prompt, which tools were called, and whether the response came from memory or a model.
          </p>
          <div className="flex flex-wrap gap-3">
            {["Auth Verified", "Cache Checked", "Tools Matched", "Response Generated"].map((step) => (
              <div key={step} className="flex items-center gap-2 px-3 py-2 rounded-xl bg-white/[0.02] border border-white/5">
                <CheckCircle2 size={13} className="text-memzent-accent" />
                <span className="text-[10px] font-black uppercase text-white/40">{step}</span>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="pt-10 border-t border-white/5">
        <div className="p-8 rounded-2xl bg-gradient-to-br from-memzent-purple/10 to-transparent border border-memzent-purple/20 flex flex-col items-center text-center gap-5">
          <Zap size={28} className="text-memzent-purple animate-pulse" />
          <h3 className="text-xl font-black uppercase tracking-tighter">Ready to go deeper?</h3>
          <p className="text-sm text-white/40 max-w-md font-bold leading-relaxed">
            Learn how to connect your own tools, pick specific AI models per request, and manage team permissions.
          </p>
          <div className="flex flex-wrap items-center gap-3">
            <Link
              href="/docs/first-request"
              className="flex items-center gap-2 px-5 py-3 rounded-xl bg-[#00FFCC] text-black text-xs font-black uppercase tracking-widest hover:scale-105 transition-all whitespace-nowrap shadow-[0_0_20px_rgba(129,231,226,0.15)]"
            >
              Explore Models <ArrowRight size={13} />
            </Link>
            <Link href="/docs/tool-registry" className="text-xs text-white/40 font-black uppercase tracking-widest hover:text-white transition-colors">
              Connect Tools →
            </Link>
          </div>
        </div>
      </section>

      <DocsPager />
    </div>
  );
}