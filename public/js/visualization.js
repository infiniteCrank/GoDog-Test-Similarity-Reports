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

// Function to render bar chart
function renderBarChart(data) {
    const metrics = ['LCS', 'Cosine', 'Jaccard'];
    const reports = [data.lcs_report, data.cosine_report, data.jaccard_report];

    const width = 800;
    const height = 400;
    const margin = {top: 20, right: 30, bottom: 40, left: 40};

    const svg = d3.select("#reportContainer").append("svg")
        .attr("width", width)
        .attr("height", height);

    const x = d3.scaleBand()
        .domain(reports[0].comparisons.map(d => d.test_a))
        .range([margin.left, width - margin.right])
        .padding(0.1);

    const y = d3.scaleLinear()
        .domain([0, d3.max(reports[0].comparisons, d => d.similarity)]).nice()
        .range([height - margin.bottom, margin.top]);

    // Add axes
    svg.append("g")
        .attr("transform", `translate(0,${height - margin.bottom})`)
        .call(d3.axisBottom(x));

    svg.append("g")
        .attr("transform", `translate(${margin.left},0)`)
        .call(d3.axisLeft(y));

    // Drawing bars for each metric
    metrics.forEach((metric, idx) => {
        svg.selectAll(`.bar-${idx}`)
            .data(reports[idx].comparisons)
            .enter().append("rect")
            .attr("class", `bar-${idx}`)
            .attr("x", d => x(d.test_a))
            .attr("y", d => y(d.similarity))
            .attr("width", x.bandwidth())
            .attr("height", d => y(0) - y(d.similarity))
            .attr("fill", d3.schemeCategory10[idx]);
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
