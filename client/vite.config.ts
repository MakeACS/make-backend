import { defineConfig } from 'vite';
import eslint from "vite-plugin-eslint2";
import react from "@vitejs/plugin-react"

export default defineConfig(({ command }) => {
  return {
    plugins: [
      react(),
      eslint({
        dev: true,
        build: true, // lint on build
        cache: true, // cache lints
        emitWarning: true,
        emitError: true,
        
      }),
    ],
    base: "/app/",
    build: {
      minify: false,
      outDir: 'build',
    },
    server: {
      port: 3001, // dont conflict with dev server which is on 3000
    },
  }
});
