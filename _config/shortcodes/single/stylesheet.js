import njk from "../../libraries/nunjucks.js";
import {sassTarget} from "../../_lib/sass.js";

export default function () {
    return njk.renderString(
        '<link rel="stylesheet" href="{{ path }}">',
        {
            path: sassTarget(),
        },
    );
}