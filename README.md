# contextual

CLI tool for fetching Jira issues, Confluence pages, and generic web pages and assembling them into **AI-ready context blocks** (`context.md`).

**The Problem:** Managing LLM context windows is hard. You want to ask an AI about a feature, but the context is scattered across three Jira tickets, a Confluence doc, and a random web page.
**The Solution:** `contextual` features **priority-driven spidering** and **context size truncation** to ensure only the most relevant, nearest content fits into your LLM's context window. Stop manually copy-pasting!

## Getting Started in 3 Steps

### Step 1: Install or Build

**Option A: Quick Install (Recommended)**
The fastest way to install `contextual` on macOS, Linux, or in automated AI agent environments is via our one-line installation script:
```bash
curl -sSfL https://raw.githubusercontent.com/jefflunt/contextual/main/install.sh | bash
```

**Option B: Build from Source**
If you prefer to clone and build locally, we recommend running the full build, test, and install suite to ensure everything works smoothly on your machine:
```bash
git clone https://github.com/jefflunt/contextual.git
cd contextual
./script/build-test-install
```

### Step 2: Configure

`contextual` requires a configuration file located at `~/.contextual/config.yml`.

Create the file with your Atlassian credentials to allow `contextual` to fetch your private issues and docs. You can generate an Atlassian API token at [id.atlassian.com/manage-profile/security/api-tokens](https://id.atlassian.com/manage-profile/security/api-tokens).

Here is a minimal configuration to get you started:

```yaml
# ~/.contextual/config.yml
max_context_length: 64000 # Max size (in bytes) of the output. Ensures you don't blow out your LLM context window.
spider:
  max_hops: 2 # How deep to follow links? (e.g., 2 hops = Epic -> Task -> Subtask)
atlassian:
  host: "your-company.atlassian.net"
  username: "your.email@example.com"
  token: "YOUR_API_TOKEN"
```
*(See `contextual.config.example.yml` in this repository for the complete list of options).*

### Step 3: Run & Generate Context

Now you're ready to fetch! You can pass multiple Jira issues, Confluence pages, or generic URLs at once. Add `--verbose` (`-v`) or `--progress` (`-p`) to see the spidering graph in action.

```bash
contextual -v -p CTX-123 https://example.com/docs/api
```

**What happens here?**
1. `contextual` fetches the initial items you requested.
2. **Spidering:** It automatically finds linked items (like child Tasks, related Epics, or linked Confluence pages) up to your configured `spider.max_hops`.
3. It orders all discovered items by priority/relevance.
4. **Truncation:** It generates a cleanly formatted `context.md` file in your current directory, automatically truncating the bottom-most (least relevant) content if the total size exceeds your `max_context_length`.

Open the resulting `context.md` file, and you'll see a perfectly formatted Markdown block ready to be passed to your AI agent or LLM interface!

## Usage for AI Agents

`contextual` is designed to be highly scriptable by AI coding agents. Agents can use the `plan` feature to generate an execution plan file based on context:

```bash
contextual plan CTX-1234
```

For more deep-dive documentation aimed specifically at AI agents (architecture, patterns), see [`agent_docs/README.md`](agent_docs/README.md).
