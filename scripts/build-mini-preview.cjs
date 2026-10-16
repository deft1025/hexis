const fs = require('node:fs');
const path = require('node:path');
const root = path.resolve(__dirname, '..');
const native = path.join(root, 'miniprogram');
const preview = path.join(root, 'internal/web/static/mini');
fs.mkdirSync(preview, { recursive: true });
for (const mode of ['home', 'champions', 'augments', 'items', 'mine', 'detail', 'search', 'guides']) {
  const dir = path.join(native, 'pages', mode);
  fs.mkdirSync(dir, { recursive: true });
  fs.writeFileSync(path.join(dir, 'index.js'), `Page(require('../../lib/page')('${mode}'));\n`);
  fs.writeFileSync(path.join(dir, 'index.wxml'), '<include src="../../lib/views.wxml" />\n');
  fs.writeFileSync(path.join(dir, 'index.json'), '{ "usingComponents": {} }\n');
}
fs.copyFileSync(path.join(native, 'lib/core.js'), path.join(preview, 'core.js'));
const css = fs.readFileSync(path.join(native, 'app.wxss'), 'utf8').replace(/(^|\n)page\{/, '$1body{');
fs.writeFileSync(path.join(preview, 'native.css'), css.replace(/(^|})([^{}]+)\{/g, (_, start, selector) => start + selector.replace(/(?<![\w.-])(view|text|image)(?![\w-])/g, tag => ({ view: 'div', text: 'span', image: 'img' })[tag]) + '{'));
if (fs.existsSync(path.join(native, 'assets'))) fs.cpSync(path.join(native, 'assets'), path.join(preview, 'assets'), { recursive: true });
console.log('Generated native page wrappers and synchronized browser preview assets.');
