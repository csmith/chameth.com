export default function(collectionApi) {
    return collectionApi.getFilteredByTag("snippets").sort(function(a, b) {
        if (a.data.group === b.data.group) {
            return a.data.title.localeCompare(b.data.title);
        } else {
            return a.data.group.localeCompare(b.data.group);
        }
    });
};