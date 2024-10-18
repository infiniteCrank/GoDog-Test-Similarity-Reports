document.getElementById('optimizeDirectoryButton').addEventListener('click', function() {
    const directory = document.getElementById('directoryInput').value;
    optimizeFeatureFiles(directory);
});

function optimizeFeatureFiles(directory) {
    const url = `/api/optimize-directory?directory=${encodeURIComponent(directory)}`;

    fetch(url)
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to optimize feature files');
            }
            return response.text(); // Get the optimized feature file content as text
        })
        .then(optimizedContent => {
            // Create a download link for the optimized file
            const blob = new Blob([optimizedContent], { type: 'text/plain' });
            const url = URL.createObjectURL(blob);
            const downloadLink = document.getElementById('downloadLink');
            downloadLink.href = url;
            downloadLink.download = 'optimized_features.feature'; // Set the name of the file
            downloadLink.style.display = 'block'; // Show the download link
            downloadLink.textContent = 'Download Optimized Feature File'; // Set link text
        })
        .catch(error => console.error('Error:', error));
}
