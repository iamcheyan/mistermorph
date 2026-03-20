import esbuild from "esbuild";

await esbuild.build({
  entryPoints: ["src/index.js"],
  outdir: "dist",
  format: "esm",
  bundle: true,
  splitting: true,
  sourcemap: false,
  minify: false,
  target: ["es2020"],
  platform: "browser",
  logLevel: "info",
  loader: {
    ".css": "css",
    ".woff": "file",
    ".woff2": "file",
    ".ttf": "file",
    ".eot": "file",
    ".svg": "file",
    ".png": "file",
    ".jpg": "file",
    ".jpeg": "file",
    ".gif": "file",
  },
  assetNames: "assets/[name]-[hash]",
});
