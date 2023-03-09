module.exports = {
  docsSidebar: [
    'introduction',
    'installation',
    'configuration',
    {
      type: 'category',
      label: 'Tour',
      items: [
        "tour/introduction",
        "tour/my-first-asset",
        "tour/querying-assets",
        "tour/asset-lineage"
      ]
    },
    {
      type: "category",
      label: "Guides",
      items: [
        "guides/ingestion",
        "guides/querying",
        "guides/starring",
        "guides/tagging",
        "guides/discussion",
      ],
    },
    {
      type: "category",
      label: "Concepts",
      items: [
        "concepts/overview",
        "concepts/asset",
        "concepts/type",
        "concepts/user",
        "concepts/architecture",
        "concepts/internals",
      ],
    },
    {
      type: "category",
      label: "Reference",
      items: [
        "reference/api",
        "reference/cli",
        "reference/configuration",
      ],
    },
    {
      type: "category",
      label: "Contribute",
      items: [
        "contribute/contributing",
        "contribute/development-guide",
      ],
    },
    'roadmap',
  ],
};