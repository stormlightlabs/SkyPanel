import { themes as prismThemes } from "prism-react-renderer";
import type { Config } from "@docusaurus/types";
import type * as Preset from "@docusaurus/preset-classic";

// This runs in Node.js - Don't use client-side code here (browser APIs, JSX...)
const config: Config = {
  title: "SkyPanel",
  tagline: "Supercharge your BlueSky/ATProto experience",
  favicon: "img/favicon.ico",
  // Future flags, see https://docusaurus.io/docs/api/docusaurus-config#future
  future: { v4: true },
  url: "https://stormlightlabs.github.io",
  baseUrl: "/skypanel/",
  organizationName: "stormlightlabs",
  projectName: "skypanel",
  onBrokenLinks: "throw",
  i18n: { defaultLocale: "en", locales: ["en"] },
  presets: [
    [
      "classic",
      { docs: { sidebarPath: "./sidebars.ts" }, theme: { customCss: "./src/css/custom.css" } } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    image: "img/docusaurus-social-card.jpg",
    colorMode: { respectPrefersColorScheme: true },
    navbar: {
      title: "My Site",
      logo: { alt: "My Site Logo", src: "img/logo.svg" },
      items: [
        { type: "docSidebar", sidebarId: "userManual", position: "left", label: "Usage" },
        { href: "https://github.com/stormlightlabs/skypanel", label: "GitHub", position: "right" },
      ],
    },
    footer: {
      style: "dark",
      links: [
        { title: "Docs", items: [{ label: "Tutorial", to: "/docs/intro" }] },
        {
          title: "Community",
          items: [
            { label: "Stack Overflow", href: "https://stackoverflow.com/questions/tagged/docusaurus" },
            { label: "Discord", href: "https://discordapp.com/invite/docusaurus" },
            { label: "BlueSky", href: "https://bsky.app/desertthunder.dev" },
          ],
        },
        { title: "More", items: [{ label: "GitHub", href: "https://github.com/stormlightlabs/skypanel" }] },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} Stormlight Labs, LLC. Made with ⚡️ in Austin, TX.`,
    },
    prism: { theme: prismThemes.github, darkTheme: prismThemes.gruvboxMaterialDark },
  } satisfies Preset.ThemeConfig,
};

export default config;
