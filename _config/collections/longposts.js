export default function(collectionApi) {
    return collectionApi.getFilteredByTag('posts').filter((a) => a.data.format === 'long');
};