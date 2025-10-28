import { fileURLToPath } from "node:url";
import { defineConfig } from "wxt";
import tailwindcss from "@tailwindcss/vite";

const resolveDir = (dir: string) => fileURLToPath(new URL(dir, import.meta.url));

const alias = {
  $assets: resolveDir("./src/assets"),
  $entrypoints: resolveDir("./src/entrypoints"),
  $lib: resolveDir("./src/lib"),
  $styles: resolveDir("./src/styles"),
};

// See https://wxt.dev/api/config.html
export default defineConfig({
  srcDir: "src",
  modules: ["@wxt-dev/module-svelte", "@wxt-dev/auto-icons"],
  vite: () => ({ plugins: [tailwindcss()], resolve: { alias } }),
  manifest: {
    action: {
      default_title: "SkyPanel",
    },
    permissions: ["sidePanel"],
  },
});
