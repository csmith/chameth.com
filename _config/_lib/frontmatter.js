import matter from 'gray-matter';
import fs from "fs";

export const readFrontMatter = (f) => matter(fs.readFileSync(f)).data;
