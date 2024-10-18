document.getElementById('fetchReports').addEventListener('click', function() {
    const directory = document.getElementById('directoryInput').value;
    fetchSimilarityReports(directory);
});

document.getElementById('fetchJourney').addEventListener('click', function() {
    const directory = document.getElementById('directoryInput2').value;
    fetchTestJourneys(directory);
});

function fetchSimilarityReports(directory) {
    let url = '/api/similarity-reports';
    if (directory) {
        url += `?directory=${encodeURIComponent(directory)}`;
    }

    fetch(url)
        .then(response => response.json())
        .then(data => {
            //renderSimilarityReports(data);
            renderForceDirectedGraph(data);
        })
        .catch(error => console.error('Error fetching reports:', error));
}

function fetchTestJourneys(directory) {
    let url = '/api/test-journeys';
    if (directory) {
        url += `?directory=${encodeURIComponent(directory)}`;
    }

    fetch(url)
        .then(response => response.json())
        .then(data => {
            renderHierarchicalChart(data); // Render the hierarchical chart with the journey data
        })
        .catch(error => console.error('Error fetching test journeys:', error));
}

// function renderSimilarityReports(data) {
//     const reportDiv = document.getElementById('similarityReport');
//     reportDiv.innerHTML = '';

//     const reportHTML = `
//         <h2>Similarity Reports</h2>
//         <h3>LCS Report</h3>
//         <pre>${JSON.stringify(data.lcs_report, null, 2)}</pre>
//         <h3>Cosine Report</h3>
//         <pre>${JSON.stringify(data.cosine_report, null, 2)}</pre>
//         <h3>Jaccard Report</h3>
//         <pre>${JSON.stringify(data.jaccard_report, null, 2)}</pre>
//     `;
//     reportDiv.innerHTML = reportHTML;
// }

function renderForceDirectedGraph(data) {
    const width = 800;
    const height = 600;

    // Clear previous graph if exists
    d3.select("#similarityReport").selectAll("svg").remove();

    // Create svg element
    const svg = d3.select("#similarityReport")
        .append("svg")
        .attr("width", width)
        .attr("height", height);

    // Prepare the nodes and links from the similarity data
    const nodes = [];
    const links = [];

    // Extract nodes and create links for LCS, Cosine, Jaccard
    const metrics = ['lcs_report', 'cosine_report', 'jaccard_report'];
    
    metrics.forEach(metric => {
        data[metric].comparisons.forEach(comparison => {
            // Add nodes if they don't already exist
            const testA = { id: comparison.testA };
            const testB = { id: comparison.testB };
            if (!nodes.find(node => node.id === testA.id)) nodes.push(testA);
            if (!nodes.find(node => node.id === testB.id)) nodes.push(testB);

            // Add a link with a weight based on similarity
            links.push({ source: testA.id, target: testB.id, weight: comparison.similarity });
        });
    });

    // Create a simulation force layout
    const simulation = d3.forceSimulation(nodes)
        .force("link", d3.forceLink().id(d => d.id).distance(100))
        .force("charge", d3.forceManyBody().strength(-300))
        .force("center", d3.forceCenter(width / 2, height / 2));

    // Add links
    const link = svg.append("g")
        .attr("class", "links")
        .selectAll("line")
        .data(links)
        .enter().append("line")
        .attr("stroke-width", d => Math.sqrt(d.weight))
        .attr("stroke", "lightgray");

    // Add nodes
    const node = svg.append("g")
        .attr("class", "nodes")
        .selectAll("circle")
        .data(nodes)
        .enter().append("circle")
        .attr("r", 5)
        .attr("fill", "#69b3a2")
        .call(d3.drag()
            .on("start", dragstarted)
            .on("drag", dragged)
            .on("end", dragended));

    // Add titles for the nodes
    node.append("title")
        .text(d => d.id);

    simulation
        .nodes(nodes)
        .on("tick", ticked);

    simulation.force("link").links(links);

    function ticked() {
        link
        .attr("x1", d => {
            const sourceNode = nodes.find(node => node.id === d.source);
            return sourceNode ? sourceNode.x : 0;
        })
        .attr("y1", d => {
            const sourceNode = nodes.find(node => node.id === d.source);
            return sourceNode ? sourceNode.y : 0;
        })
        .attr("x2", d => {
            const targetNode = nodes.find(node => node.id === d.target);
            return targetNode ? targetNode.x : 0;
        })
        .attr("y2", d => {
            const targetNode = nodes.find(node => node.id === d.target);
            return targetNode ? targetNode.y : 0;
        });

    node
        .attr("cx", d => d.x)
        .attr("cy", d => d.y);
}

// Dragging functionality for nodes
function dragstarted(event, d) {
    if (!event.active) simulation.alphaTarget(0.3).restart();
    d.fx = d.x;
    d.fy = d.y;
}

function dragged(event, d) {
    d.fx = event.x;
    d.fy = event.y;
}

function dragended(event, d) {
    if (!event.active) simulation.alphaTarget(0);
    d.fx = null;
    d.fy = null;
}
}


function renderHierarchicalChart(data) {
    const hierarchyData = {
        name: "Test Journeys",
        children: data.children // This should reflect the structure from your API response
    };

    const width = 600;
    const height = 400;

    const svg = d3.select("#hierarchicalChart")
        .append("svg")
        .attr("width", width)
        .attr("height", height);

    const root = d3.hierarchy(hierarchyData);
    const treeLayout = d3.tree().size([height, width - 160]);
    treeLayout(root);

    // Links
    svg.selectAll('.link')
        .data(root.links())
        .enter()
        .append('line')
        .attr('class', 'link')
        .attr('x1', d => d.source.y)
        .attr('y1', d => d.source.x)
        .attr('x2', d => d.target.y)
        .attr('y2', d => d.target.x)
        .attr('stroke', '#ccc');

    // Nodes
    const nodes = svg.selectAll('.node')
        .data(root.descendants())
        .enter()
        .append('g')
        .attr('class', d => 'node' + (d.children ? ' node--internal' : ' node--leaf'))
        .attr('transform', d => `translate(${d.y},${d.x})`);

    nodes.append('circle')
        .attr('r', 4.5)
        .attr('fill', '#69b3a2');

    nodes.append('text')
        .attr('dy', 3)
        .attr('x', d => d.children ? -8 : 8)
        .style('text-anchor', d => d.children ? 'end' : 'start')
        .text(d => d.data.name);
}

