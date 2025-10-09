import fs from "fs";

export default (file) =>
    fs.existsSync(file);