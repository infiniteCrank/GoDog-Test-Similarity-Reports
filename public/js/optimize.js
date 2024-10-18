document.getElementById('optimizeForm').addEventListener('submit', function(event) {
    event.preventDefault(); // Prevent form from submitting the default way

    const formData = new FormData(this);
    
    // Send the form data to the API
    fetch('/api/optimize-feature', {
        method: 'POST',
        body: formData
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('Failed to optimize the feature file');
        }
        return response.text(); // Get the optimized file content as text
    })
    .then(optimizedContent => {
        // Display the optimized content in the pre tag
        document.getElementById('optimizedContent').textContent = optimizedContent;

        // Create a download link for the optimized file
        const blob = new Blob([optimizedContent], { type: 'text/plain' });
        const url = URL.createObjectURL(blob);
        const downloadLink = document.getElementById('downloadLink');
        downloadLink.href = url;
        downloadLink.download = 'optimized.feature'; // Setting the name of the optimized file
        downloadLink.style.display = 'block'; // Show the download link
    })
    .catch(error => console.error('Error:', error));
});
