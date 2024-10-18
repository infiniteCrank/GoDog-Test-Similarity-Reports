document.getElementById('fetchReports').addEventListener('click', function() {
    const selectedChartType = document.getElementById('chartType').value;
    fetchSimilarityReports(selectedChartType);
});

function fetchSimilarityReports(chartType) {
    const url = '/api/similarity-reports'; // Adjust if needed

    fetch(url)
        .then(response => response.json())
        .then(data => {
            // Clear previous content
            d3.select("#reportContainer").selectAll("*").remove();

            if (chartType === 'bar') {
                renderBarChart(data);
            } else if (chartType === 'heatmap') {
                renderHeatmap(data);
            } else if (chartType === 'radar') {
                renderRadarChart(data);
            }
        })
        .catch(error => console.error('Error fetching reports:', error));
}

function renderBarChart(data) {
    // Define the metrics we will visualize
    const metrics = ['lcs_report', 'cosine_report', 'jaccard_report'];

    // Prepare the reports for each metric
    const reports = [data.lcs_report, data.cosine_report, data.jaccard_report];

    // Set the dimensions of the SVG for the bar chart
    const width = 800;
    const height = 400;
    const margin = { top: 20, right: 30, bottom: 100, left: 40 }; // Added margin for aesthetics

    // Append an SVG to the report container for visual representation
    const svg = d3.select("#reportContainer").append("svg")
        .attr("width", width)
        .attr("height", height);

    // Create a scale for the x-axis using band scale
    const x = d3.scaleBand()
        .domain(reports[0].comparisons.map(d => d.test_a)) // Use test names as the domain
        .range([margin.left, width - margin.right]) // Define range based on SVG dimensions
        .padding(0.1); // Space between bars

    // Create a scale for the y-axis using linear scale
    const y = d3.scaleLinear()
        .domain([0, d3.max(reports, report => d3.max(report.comparisons, d => d.similarity))]).nice()
        .range([height - margin.bottom, margin.top]); // Y values range from top to bottom of the SVG

    // Add x-axis to the SVG
    svg.append("g")
        .attr("transform", `translate(0,${height - margin.bottom})`) // Position it at the bottom
        .call(d3.axisBottom(x)) // Call the axisBottom function to draw the x-axis and its ticks
        .selectAll("text")
        .attr("transform", "rotate(-45)") // Rotate labels to a vertical position for better readability
        .attr("dx", "-0.8em") // Adjust horizontal position to fit
        .attr("dy", ".15em") // Adjust vertical position to reduce collision with x-axis
        .style("text-anchor", "end"); // Anchor the text at the end for correct alignment

    // Add y-axis to the SVG
    svg.append("g")
        .attr("transform", `translate(${margin.left},0)`) // Position it at the left margin
        .call(d3.axisLeft(y)); // Call the axisLeft function to draw the y-axis and its ticks

    // Drawing bars for each metric
    const barWidth = x.bandwidth() / metrics.length; // Calculate the width of each bar based on the number of metrics

    metrics.forEach((metric, idx) => { // Loop through each metric
        svg.selectAll(`.bar-${idx}`) // Select elements with class related to the current metric
            .data(reports[idx].comparisons) // Bind data for the current metric
            .enter().append("rect") // Enter selection for appending rectangles (bars)
            .attr("class", `bar-${idx}`) // Set the class for the bars based on their metric
            .attr("x", d => x(d.test_a) + idx * barWidth) // Position the bars on the x-axis, offset for multiple metrics
            .attr("y", d => y(d.similarity)) // Set the height of the bar based on similarity
            .attr("width", barWidth - 1) // Set the width of the bar and include spacing between bars
            .attr("height", d => y(0) - y(d.similarity)) // Set the height of the bar, calculating from the y-scale
            .attr("fill", d3.schemeCategory10[idx]); // Use a color scheme for the bars based on the metric index
    });

    // Render the legend
    renderLegend(metrics);
}

// Function to render the legend
function renderLegend(metrics) {
    const legendContainer = d3.select("#legendContainer");
    legendContainer.selectAll("*").remove(); // Clear existing legends

    // Create a legend item for each metric
    metrics.forEach((metric, idx) => {
        const legendItem = legendContainer.append("div").attr("class", "legend");
        
        legendItem.append("div")
            .attr("class", "square")
            .style("background-color", d3.schemeCategory10[idx]); // Use the same color as the bars

        legendItem.append("span")
            .text(metric) // Display metric name next to the color square
            .style("font-size", "16px") // Optional: Style the font size
            .style("margin-right", "10px"); // Optional: Add spacing
    });
}

// Function to render heatmap
function renderHeatmap(data) {
    const reports = [data.lcs_report, data.cosine_report, data.jaccard_report];

    // Prepare data for heatmap
    const heatmapData = [];
    reports.forEach(report => {
        report.comparisons.forEach(comp => {
            heatmapData.push({
                testA: comp.testA,
                testB: comp.testB,
                similarity: comp.similarity,
                metric: report.similarity_type
            });
        });
    });

    const width = 600;
    const height = 400;
    const margin = { top: 20, right: 30, bottom: 40, left: 40 };

    const svg = d3.select("#reportContainer").append("svg")
        .attr("width", width)
        .attr("height", height);

    const x = d3.scaleBand()
        .domain(heatmapData.map(d => d.testA))
        .range([margin.left, width - margin.right])
        .padding(0.01);

    const y = d3.scaleBand()
        .domain(heatmapData.map(d => d.testB))
        .range([height - margin.bottom, margin.top])
        .padding(0.01);

    // Draw heatmap squares
    svg.selectAll("rect")
        .data(heatmapData)
        .enter().append("rect")
        .attr("x", d => x(d.testA))
        .attr("y", d => y(d.testB))
        .attr("width", x.bandwidth())
        .attr("height", y.bandwidth())
        .style("fill", d => {
            const colorScale = d3.scaleSequential(d3.interpolateRdYlBu).domain([0, 1]);
            return colorScale(d.similarity);
        });

    // Add axes
    svg.append("g")
        .attr("transform", `translate(0,${height - margin.bottom})`)
        .call(d3.axisBottom(x));

    svg.append("g")
    .attr("transform", `translate(${margin.left},0)`)
    .call(d3.axisLeft(y));
}

// Function to render radar chart
function renderRadarChart(data) {
const reports = [data.lcs_report, data.cosine_report, data.jaccard_report];

// Prepare nodes based on reports
const nodes = reports.map(report => ({
    name: report.similarity_type,
    values: report.comparisons.map(comp => ({
        name: comp.test_a,
        value: comp.similarity
    }))
}));

const width = 600;
const height = 600;

// Create the SVG element for radar chart
const svg = d3.select("#reportContainer").append("svg")
    .attr("width", width)
    .attr("height", height);

const angleSlice = (Math.PI * 2) / nodes[0].values.length;

// Set the radius for the radar chart
const radius = Math.min(width, height) / 2;

// Create a scale for radius values
const radarLine = d3.lineRadial()
    .angle((d, i) => (i * 2 * Math.PI) / nodes[0].values.length)
    .radius(d => d.value * (radius / 1)); // Scale the radius value

// Draw each radar for the metrics
nodes.forEach((metric, idx) => {
    svg.append("path")
        .datum(metric.values)
        .attr("d", radarLine)
        .attr("transform", `translate(${width / 2}, ${height / 2})`)
        .attr("fill", d3.schemeCategory10[idx])
        .attr("stroke", "black")
        .attr("opacity", 0.7);
});

// Draw axis lines and labels
nodes[0].values.forEach((d, i) => {
    const angle = angleSlice * i;
    const x = Math.cos(angle) * radius + width / 2;
    const y = Math.sin(angle) * radius + height / 2;

    svg.append("line")
        .attr("x1", width / 2)
        .attr("y1", height / 2)
        .attr("x2", x)
        .attr("y2", y)
        .attr("stroke", "gray");

    svg.append("text")
        .attr("x", x)
        .attr("y", y)
        .attr("dy", -5) // Slightly above the axis
        .attr("text-anchor", "middle")
        .text(d.name);
});
}
