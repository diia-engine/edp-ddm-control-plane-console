document.addEventListener('DOMContentLoaded', () => {
    const targetElement = document.getElementById('changes');
    const configuration = { drawFileList: false, matching: 'lines', highlight: true, outputFormat: 'side-by-side',
        rawTemplates: {"tag-file-renamed": ""}};
    const diff2htmlUi = new Diff2HtmlUI(targetElement, diffString, configuration);
    diff2htmlUi.draw();
    diff2htmlUi.highlightCode();
});