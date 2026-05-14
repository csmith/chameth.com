<?xml version="1.0" encoding="utf-8"?>
<xsl:stylesheet version="3.0"
                xmlns:xsl="http://www.w3.org/1999/XSL/Transform"
                xmlns:atom="http://www.w3.org/2005/Atom">
    <xsl:output method="html" version="1.0" encoding="UTF-8" indent="yes"/>
    <xsl:template match="/">
        <html xmlns="http://www.w3.org/1999/xhtml" lang="en">
            <head>
                <title>
                    Feed:
                    <xsl:value-of select="/atom:feed/atom:title"/>
                </title>
                <link rel="stylesheet" href="/feeds.css"/>
            </head>
            <body>
                <h1>Feed preview</h1>
                <a href="/">← Back to chameth.com</a>
                <p>
                    This is an Atom feed. You can add it to a feed reader
                    application to read new content when it's published. Visit
                    <a href="https://aboutfeeds.com">About Feeds</a>
                    to learn more and get started. It's free!
                </p>
                <h2>Content</h2>
                <xsl:for-each select="/atom:feed/atom:entry">
                    <section>
                        <a>
                            <xsl:attribute name="href">
                                <xsl:value-of select="atom:link/@href"/>
                            </xsl:attribute>
                            <xsl:value-of select="atom:title"/>
                        </a>
                        <xsl:value-of select="atom:summary"/>
                        Posted:
                        <xsl:value-of select="substring(atom:updated, 0, 11)"/>
                    </section>
                </xsl:for-each>
            </body>
        </html>
    </xsl:template>
</xsl:stylesheet>