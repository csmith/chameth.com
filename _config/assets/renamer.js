import path from "path";
import fs from "fs";
import {readFrontMatter} from "../_lib/frontmatter.js";

export default function (original) {
    // See if there's an index.md alongside the asset, if so figure out the permalink and put the file there.
    // This emulates Hugo's "page bundles" behaviour.
    const content = path.join(path.dirname(original), 'index.md');
    if (fs.existsSync(content)) {
        const data = readFrontMatter(content);
        if (data && data.permalink) {
            return path.join(data.permalink, path.basename(original));
        }
    }

    // Dump content in /static at the root of the output.
    if (original.startsWith('static/')) {
        return original.replace('static/', '');
    }
    return original;
};