import Nunjucks from 'nunjucks';

const njk = new Nunjucks.Environment(
    new Nunjucks.FileSystemLoader('_includes')
);

export default njk;
