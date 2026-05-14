/**
 * 下载文件
 * @param url 文件下载地址
 */
export function downloadFile(url: string) {
    // 使用隐藏的 iframe 下载，避免页面闪烁
    const iframe = document.createElement('iframe');
    iframe.style.display = 'none';
    iframe.src = url;
    document.body.appendChild(iframe);

    // 1秒后移除 iframe
    setTimeout(() => {
        document.body.removeChild(iframe);
    }, 1000);
}
