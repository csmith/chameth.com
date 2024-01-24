import fs from "fs";
import path from "path";
import crypto from "crypto";

let hash = '';

export const sassHash = () => {
    if (hash !== '') {
        return hash;
    }

    hash = crypto.createHash('sha256').update(
        fs.readdirSync('static', {recursive: true})
            .filter(p => path.extname(p) === '.scss')
            .sort()
            .map(p => fs.readFileSync(path.join('static', p)))
            .join()
    ).digest('hex').substring(0, 8);
    return hash;
}

export const sassTarget = () => `global-${sassHash()}.css`