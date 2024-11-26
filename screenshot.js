const puppeteer = require('puppeteer');

function sleep(ms) {
    ms = (ms) ? ms : 0;
    return new Promise(resolve => {
        setTimeout(resolve, ms);
    });
}

process.on('uncaughtException', (error) => {
    console.error(error);
    process.exit(1);
});

process.on('unhandledRejection', (reason, p) => {
    console.error(reason, p);
    process.exit(1);
});

const url = process.argv[1];
const filename = process.argv[2];

const width = 1920;
const height = 1080;
const delay = 1000;
const isMobile = false;

(async () => {

    const browser = await puppeteer.launch({
        args: [
            '--no-sandbox',
            '--disable-setuid-sandbox',
            '--disable-web-security'
        ]
    });

    const page = await browser.newPage();

    page.setViewport({
        width,
        height,
        isMobile
    });

    await page.goto(url, {waitUntil: 'networkidle2'});

    // Fix Some URLs to be relative not absolute
    await page.evaluate(() => {
        items = document.querySelectorAll('img');
        items.forEach((item) => {
            item.src = item.src.replace('file:///', './');
        });

        items = document.querySelectorAll('link');
        items.forEach((item) => {
            item.href = item.href.replace('file:///', './');
        });

        items = document.querySelectorAll('script');
        items.forEach((item) => {
            item.src = item.src.replace('file:///', './');
        });
    })

    await sleep(delay);

    await page.screenshot({path: filename, fullPage: false});

    browser.close();
})();