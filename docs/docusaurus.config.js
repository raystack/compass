const lightCodeTheme = require('prism-react-renderer/themes/dracula');
const darkCodeTheme = require('prism-react-renderer/themes/dracula');

// With JSDoc @type annotations, IDEs can provide config autocompletion
/** @type {import('@docusaurus/types').DocusaurusConfig} */
(module.exports = {
  title: 'Compass',
  tagline: 'Data Catalog',
  url: 'https://raystack.github.io/',
  baseUrl: '/compass/',
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',
  favicon: 'assets/favicon.ico',
  organizationName: 'Raystack',
  projectName: 'compass',

  presets: [
    [
      '@docusaurus/preset-classic',
      ({
        gtag: {
          trackingID: 'G-XXX',
        },
        docs: {
          routeBasePath: '/',
          sidebarPath: require.resolve('./sidebars.js'),
          editUrl: 'https://github.com/raystack/compass/edit/master/docs/',
          sidebarCollapsed: true,
        },
        blog: false,
        theme: {
          customCss: [
            require.resolve('./src/css/theme.css'),
            require.resolve('./src/css/custom.css'),
            require.resolve('./src/css/icons.css'),
          ],
        },
      })
    ],
  ],

  themeConfig:
    ({
      colorMode: {
        defaultMode: 'light',
        respectPrefersColorScheme: true,
      },
      navbar: {
        title: 'Compass',
        logo: { src: 'assets/logo.svg', },
        hideOnScroll: true,
        items: [
          {
            type: 'doc',
            docId: 'introduction',
            position: 'left',
            label: 'Docs',
          },
          { to: '/help', label: 'Help', position: 'left' },
          {
            href: 'https://bit.ly/2RzPbtn',
            position: 'right',
            className: 'header-slack-link',
          },
          {
            href: 'https://github.com/raystack/compass',
            className: 'navbar-item-github',
            position: 'right',
          },
        ],
      },
      footer: {
        style: 'light',
        links: [
          {
            title: 'Products',
            items: [
              { label: 'Optimus', href: 'https://github.com/raystack/optimus' },
              { label: 'Firehose', href: 'https://github.com/raystack/firehose' },
              { label: 'Raccoon', href: 'https://github.com/raystack/raccoon' },
              { label: 'Dagger', href: 'https://raystack.github.io/dagger/' },
            ],
          },
          {
            title: 'Resources',
            items: [
              { label: 'Docs', to: '/' },
              { label: 'Help', to: '/help', },
            ],
          },
          {
            title: 'Community',
            items: [
              { label: 'Slack', href: 'https://bit.ly/2RzPbtn' },
              { label: 'GitHub', href: 'https://github.com/raystack/compass' }
            ],
          },
        ],
      },
      prism: {
        theme: lightCodeTheme,
        darkTheme: darkCodeTheme,
      },
      announcementBar: {
        id: 'star-repo',
        content: '⭐️ If you like Compass, give it a star on <a target="_blank" rel="noopener noreferrer" href="https://github.com/raystack/compass">GitHub</a>! ⭐',
        backgroundColor: '#222',
        textColor: '#eee',
        isCloseable: true,
      },
    }),
});
