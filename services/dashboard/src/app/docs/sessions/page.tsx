import { Database, MessageCircle, Brain, Code } from "lucide-react";
import { DocsPager } from "@/components/docs/docs-pager";
import { CodeBlock } from "@/components/docs/code-block";
import { DOCS_CONFIG } from "@/config/docs-config";

export default function SessionsPage() {
  const createSession = `curl -X POST https://${DOCS_CONFIG.domain}/v1/sessions \\
  -H "X-API-Key: memzent_YOUR_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{"title": "Research on quantum computing"}'`;

  const createResponse = `{
  "id": "sess_a1b2c3d4...",
  "session_id": "sess_a1b2c3d4..."
}`;

  const chatWithSession = `curl -X POST https://${DOCS_CONFIG.domain}/v1/chat \\
  -H "X-API-Key: memzent_YOUR_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "messages": [{"role": "user", "content": "What is quantum entanglement?"}],
    "session_id": "sess_a1b2c3d4..."
  }'`;

  const getMessages = `curl -X GET https://${DOCS_CONFIG.domain}/v1/sessions/sess_a1b2c3d4.../messages \\
  -H "X-API-Key: memzent_YOUR_KEY"`;

  const messagesResponse = `[
  { "role": "user", "content": "What is quantum entanglement?" },
  { "role": "assistant", "content": "Quantum entanglement is a phenomenon where..." },
  { "role": "user", "content": "How does it relate to teleportation?" },
  { "role": "assistant", "content": "Quantum teleportation uses entanglement to..." }
]`;

  const deleteSession = `curl -X DELETE https://${DOCS_CONFIG.domain}/v1/sessions/sess_a1b2c3d4... \\
  -H "X-API-Key: memzent_YOUR_KEY"`;

  return (
    <div className="max-w-4xl">
      <div className="flex items-center gap-3 mb-4">
        <div className="p-2 rounded-xl bg-memzent-glow/10 border border-memzent-glow/20">
          <MessageCircle size={20} className="text-memzent-glow" />
        </div>
        <h1 className="text-3xl font-black tracking-tight">Sessions & Memory</h1>
      </div>
      <p className="text-white/50 text-sm leading-relaxed mb-10">
        Sessions give your LLM interactions persistent context. Attach a session ID to your requests
        and Memzent automatically maintains conversation history and extracts long-term semantic memory.
      </p>

      {/* How it works */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Brain size={16} className="text-memzent-glow" />
          How Memory Works
        </h2>

        <div className="space-y-4">
          <div className="p-4 rounded-xl border border-white/5 bg-white/[0.02]">
            <h4 className="text-xs font-black text-memzent-glow/70 mb-2">Session Memory (Short-term)</h4>
            <p className="text-xs text-white/40">
              Full conversation history maintained per-session. Previous messages are included in LLM context
              automatically. Scoped to a single conversation thread.
            </p>
          </div>

          <div className="p-4 rounded-xl border border-white/5 bg-white/[0.02]">
            <h4 className="text-xs font-black text-memzent-glow/70 mb-2">Semantic Memory (Long-term)</h4>
            <p className="text-xs text-white/40">
              After each exchange, Memzent automatically extracts permanent facts about the user
              (tech stack, preferences, configurations) and stores them as vectors in Qdrant.
              These are recalled across all future sessions when semantically relevant.
            </p>
          </div>

          <div className="p-4 rounded-xl border border-white/5 bg-white/[0.02]">
            <h4 className="text-xs font-black text-memzent-glow/70 mb-2">Memory Recall</h4>
            <p className="text-xs text-white/40">
              On every request, the engine queries stored memories with a relevance threshold of 0.65.
              Relevant facts are injected into the LLM context, giving it knowledge of past interactions
              without the user needing to repeat themselves.
            </p>
          </div>
        </div>
      </section>

      {/* Creating Sessions */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Database size={16} className="text-memzent-glow" />
          Creating a Session
        </h2>
        <CodeBlock code={createSession} language="bash" title="Create Session" />
        <CodeBlock code={createResponse} language="json" title="Response" />
      </section>

      {/* Using Sessions */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4 flex items-center gap-2">
          <Code size={16} className="text-memzent-glow" />
          Using Sessions in Chat
        </h2>
        <p className="text-white/50 text-sm mb-4">
          Pass the <code className="text-memzent-glow/70">session_id</code> in your chat requests.
          Memzent automatically appends the conversation to the session and includes history in the LLM context.
        </p>
        <CodeBlock code={chatWithSession} language="bash" title="Chat with Session" />
      </section>

      {/* Retrieving History */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4">Retrieving Conversation History</h2>
        <CodeBlock code={getMessages} language="bash" title="Get Messages" />
        <CodeBlock code={messagesResponse} language="json" title="Response" />
      </section>

      {/* Deleting */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4">Deleting a Session</h2>
        <CodeBlock code={deleteSession} language="bash" title="Delete Session" />
        <p className="text-white/50 text-sm mt-4">
          Returns <code className="text-memzent-glow/70">{`{"status": "deleted"}`}</code> on success.
          Semantic memories extracted during the session are <strong>not</strong> deleted — they persist
          as long-term org knowledge.
        </p>
      </section>

      {/* Best practices */}
      <section className="mb-12">
        <h2 className="text-xl font-black mb-4">Best Practices</h2>
        <div className="space-y-3 text-sm text-white/50">
          <div className="flex items-start gap-3">
            <span className="text-memzent-glow font-mono text-xs mt-0.5">•</span>
            <span>Create one session per logical conversation or task to keep context focused.</span>
          </div>
          <div className="flex items-start gap-3">
            <span className="text-memzent-glow font-mono text-xs mt-0.5">•</span>
            <span>For stateless one-shot queries, omit <code className="text-memzent-glow/70">session_id</code> entirely.</span>
          </div>
          <div className="flex items-start gap-3">
            <span className="text-memzent-glow font-mono text-xs mt-0.5">•</span>
            <span>Sessions persist across requests — no need to resend full history each time.</span>
          </div>
          <div className="flex items-start gap-3">
            <span className="text-memzent-glow font-mono text-xs mt-0.5">•</span>
            <span>Semantic memory works automatically. Facts like &quot;I use PostgreSQL 16&quot; will be recalled in future sessions.</span>
          </div>
        </div>
      </section>

      <DocsPager />
    </div>
  );
}
