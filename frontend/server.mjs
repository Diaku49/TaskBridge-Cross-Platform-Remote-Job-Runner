import { createServer } from "node:http";
import { readFile } from "node:fs/promises";
import { extname, join, normalize } from "node:path";
import { fileURLToPath } from "node:url";

const root = fileURLToPath(new URL(".", import.meta.url));
const port = Number(process.env.PORT || 5173);
const host = process.env.HOST || "127.0.0.1";
const apiBase = process.env.TASKBRIDGE_API || "http://localhost:8080";
const jobTargetBase = process.env.TASKBRIDGE_JOB_TARGET || apiBase;

const contentTypes = {
  ".html": "text/html; charset=utf-8",
  ".css": "text/css; charset=utf-8",
  ".js": "text/javascript; charset=utf-8",
  ".json": "application/json; charset=utf-8",
  ".svg": "image/svg+xml",
};

const server = createServer(async (req, res) => {
  try {
    const url = new URL(req.url || "/", `http://${req.headers.host || "localhost"}`);

    if (url.pathname === "/config.js") {
      serveConfig(res);
      return;
    }

    await serveStatic(res, url.pathname);
  } catch (error) {
    res.writeHead(500, { "Content-Type": "text/plain; charset=utf-8" });
    res.end(`frontend server failed: ${error.message}`);
  }
});

server.listen(port, host, () => {
  console.log(`TaskBridge dashboard: http://${host}:${port}`);
  console.log(`API base: ${apiBase}`);
});

function serveConfig(res) {
  const body = `window.__TASKBRIDGE_CONFIG__ = ${JSON.stringify({ apiBase, jobTargetBase })};\n`;
  res.writeHead(200, {
    "Content-Type": contentTypes[".js"],
    "Cache-Control": "no-store",
  });
  res.end(body);
}

async function serveStatic(res, pathname) {
  const requested = pathname === "/" ? "/index.html" : pathname;
  const safePath = normalize(decodeURIComponent(requested)).replace(/^(\.\.[/\\])+/, "");
  const filePath = join(root, safePath);
  const extension = extname(filePath);

  try {
    const data = await readFile(filePath);
    res.writeHead(200, {
      "Content-Type": contentTypes[extension] || "application/octet-stream",
      "Cache-Control": "no-store",
    });
    res.end(data);
  } catch {
    const data = await readFile(join(root, "index.html"));
    res.writeHead(200, {
      "Content-Type": contentTypes[".html"],
      "Cache-Control": "no-store",
    });
    res.end(data);
  }
}
