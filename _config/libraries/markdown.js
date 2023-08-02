import markdownIt from 'markdown-it';
import footnote from 'markdown-it-footnote';
import prism from 'markdown-it-prism';

const md = markdownIt({html: true, typographer: true})
    .use(footnote)
    .use(prism);

md.renderer.rules.table_open = () => '<div class="table-holder"><table>';
md.renderer.rules.table_close = () => '</table></div>';

export default md;