/**
 * Lightweight Markdown renderer.
 * HTML is escaped before any Markdown processing, preventing XSS injection.
 * Only http/https URLs are rendered as anchor tags.
 */
function renderMarkdown(text) {
    if (!text) return '<span style="color:#9ca3af"><em>No description</em></span>';
    // Escape HTML to prevent XSS
    let s = text
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;');
    // Fenced code blocks
    s = s.replace(/```([\s\S]*?)```/g, function(_, code) {
        return '<pre><code>' + code.trim() + '</code></pre>';
    });
    // Inline code
    s = s.replace(/`([^`\n]+)`/g, '<code>$1</code>');
    // Headings
    s = s.replace(/^### (.+)$/gm, '<h3>$1</h3>');
    s = s.replace(/^## (.+)$/gm, '<h2>$1</h2>');
    s = s.replace(/^# (.+)$/gm, '<h1>$1</h1>');
    // Horizontal rules (processed before lists to prevent mismatching)
    s = s.replace(/^[-*_]{3,}$/gm, '<hr>');
    // Bold + italic, bold, italic (most-greedy first to avoid partial matches)
    s = s.replace(/\*\*\*(.+?)\*\*\*/g, '<strong><em>$1</em></strong>');
    s = s.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
    s = s.replace(/__(.+?)__/g, '<strong>$1</strong>');
    s = s.replace(/\*([^*\n]+)\*/g, '<em>$1</em>');
    s = s.replace(/_([^_\n]+)_/g, '<em>$1</em>');
    // Links - only allow http/https URLs
    s = s.replace(/\[([^\]]+)\]\(([^)]+)\)/g, function(_, linkText, url) {
        if (/^https?:\/\//.test(url)) {
            return '<a href="' + url + '" target="_blank" rel="noopener noreferrer">' + linkText + '</a>';
        }
        return linkText;
    });
    // Blockquotes (> was HTML-escaped to &gt;)
    s = s.replace(/^&gt;\s*(.+)$/gm, '<blockquote>$1</blockquote>');
    // Unordered lists
    s = s.replace(/((?:^[-*+] .+\n?)+)/gm, function(match) {
        const items = match.trim().split('\n')
            .map(item => '<li>' + item.replace(/^[-*+] /, '') + '</li>').join('');
        return '<ul>' + items + '</ul>';
    });
    // Ordered lists
    s = s.replace(/((?:^\d+[.)]\s.+\n?)+)/gm, function(match) {
        const items = match.trim().split('\n')
            .map(item => '<li>' + item.replace(/^\d+[.)]\s/, '') + '</li>').join('');
        return '<ol>' + items + '</ol>';
    });
    // Paragraphs (split on blank lines)
    const blocks = s.split(/\n\n+/);
    s = blocks.map(function(block) {
        block = block.trim();
        if (!block) return '';
        if (/^<(?:h[1-6]|ul|ol|pre|hr|blockquote)/.test(block)) return block;
        return '<p>' + block.replace(/\n/g, '<br>') + '</p>';
    }).filter(Boolean).join('\n');
    return s || '<span style="color:#9ca3af"><em>No description</em></span>';
}
