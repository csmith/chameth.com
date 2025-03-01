import excerpt from "./excerpt.js";

export default async (page) => {
    return /<p>([\s\S]*?)<\/p>/
        .exec(await excerpt(page))[1]
        .replaceAll(/<.*?>/g, '');
}