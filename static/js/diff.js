document.addEventListener('DOMContentLoaded', () => {
    const targetElement = document.getElementById('changes');
    const configuration = { drawFileList: false, matching: 'lines', highlight: true, outputFormat: 'side-by-side',
        rawTemplates: {"tag-file-renamed": ""}};
    let originalDiffString = JSON.parse(diffString);
    const diff2htmlUi = new Diff2HtmlUI(targetElement, originalDiffString, configuration);
    diff2htmlUi.draw();
    diff2htmlUi.highlightCode();
});