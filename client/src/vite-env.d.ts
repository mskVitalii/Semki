/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_URL: string
  readonly TEST_EMAIL?: string
  readonly TEST_PASSWORD?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
