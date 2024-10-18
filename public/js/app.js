document.getElementById('fetchReports').addEventListener('click', function() {
    const files = document.getElementById('directoryInput').files;
    // Get the directory name
    const directory = files.length > 0 ? files[0].webkitRelativePath.split('/')[0] : null; 

    fetchSimilarityReports(directory);

});

document.getElementById('fetchMergedJourneys').addEventListener('click', function() {
    const directory = document.getElementById('mergedDirectoryInput').value;
    fetchMergedTestJourneys(directory);
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

function fetchMergedTestJourneys(directory) {
    const url = `/api/merged-test-journeys?directory=${encodeURIComponent(directory)}`;

    fetch(url)
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to fetch merged test journeys');
            }
            return response.json(); // Parse the JSON response
        })
        .then(data => {
            renderMergedJourneys(data); // Call a function to render the data
        })
        .catch(error => console.error('Error fetching merged journeys:', error));
}

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

function renderMergedJourneys(data) {
    // Create a hierarchical structure from the incoming data
    const hierarchyData = {
        name: data.name, // Set the root node name
        children: data.children // Children nodes directly sourced from API response
    };

    // Set up SVG dimensions to be responsive
    const width = window.innerWidth; // Full width of the SVG based on the window's width
    const height = window.innerHeight; // Full height of the SVG based on the window's height

    // Clear any existing SVG elements in the merged journeys container
    d3.select("#mergedJourneysContainer").selectAll("svg").remove();

    // Create the SVG element for displaying the merged test journeys
    const svg = d3.select("#mergedJourneysContainer")
        .append("svg") // Append an SVG to the container
        .attr("width", width) // Set SVG width
        .attr("height", height); // Set SVG height

    // Create a D3 hierarchy of the data
    const root = d3.hierarchy(hierarchyData);
    // Define the tree layout specifying the size
    const treeLayout = d3.tree()
        .size([height - 100, width - 160]); // Adjusted size for better spacing

    // Set the separation factor for siblings in the layout
    treeLayout.separation = (a, b) => {
        return (a.parent === b.parent ? 1 : 1.5); // Define spacing between sibling nodes
    };

    // Compute the layout based on the D3 hierarchy
    treeLayout(root);

    // Draw links (connecting lines between nodes)
    const links = svg.selectAll('.link')
        .data(root.links()) // Bind data to the links
        .enter()
        .append('line') // Append lines for each link
        .attr('class', 'link') // Set class attribute for CSS styling
        .attr('x1', d => d.source.y) // Starting x position from the source node
        .attr('y1', d => d.source.x) // Starting y position from the source node
        .attr('x2', d => d.target.y) // Ending x position for the target node
        .attr('y2', d => d.target.x) // Ending y position for the target node
        .attr('stroke', '#ccc'); // Set stroke color for links

    // Draw nodes (elements representing the data points)
    const nodes = svg.selectAll('.node')
        .data(root.descendants()) // Bind data to the descendant nodes
        .enter()
        .append('g') // Append group elements for each node
        .attr('class', d => 'node' + (d.children ? ' node--internal' : ' node--leaf')) // Assign class based on whether it has children
        .attr('transform', d => `translate(${d.y},${d.x})`) // Position nodes based on layout
        .call(d3.drag() // Enable dragging for nodes
            .on("start", dragstarted)
            .on("drag", dragged)
            .on("end", dragended)
        );

    // Add circles as visual nodes
    nodes.append('circle')
        .attr('r', 5) // Set the radius of the circles
        .attr('fill', '#69b3a2'); // Fill color for nodes

    // Add text labels for each node
    nodes.append('text')
        .attr('dy', 3) // Vertical alignment
        .attr('x', d => d.children ? -8 : 8) // Adjust horizontal position based on children
        .style('text-anchor', d => d.children ? 'end' : 'start') // Set text alignment
        .text(d => d.data.name); // Display the name of the node

    // Dragging functions for interactivity
    function dragstarted(event, d) {
        d3.select(this).raise() // Raise the dragged element above others
            .classed("active", true); // Set class to indicate that we are dragging
    }

    function dragged(event, d) {
        // Update the node position on drag
        d.x = event.y; // Update the y position based on mouse movement
        d.y = event.x; // Update the x position based on mouse movement

        // Move the node visually in the chart
        d3.select(this) // Selected node's group
            .attr("transform", `translate(${d.y}, ${d.x})`);

        // Update the links' positions to reflect the moved node
        svg.selectAll('.link')
        .attr("x1", l => l.source.y) // Update x1 position for links
        .attr("y1", l => l.source.x) // Update y1 position for links
        .attr("x2", l => l.target.y) // Update x2 position for links
        .attr("y2", l => l.target.x); // Update y2 position for links
    }

    function dragended(event, d) {
        d3.select(this).classed("active", false); // Remove active class when drag ends
        // You can implement logic here to finalize node positions or allow them to be free
        // For example:
        // d.fx = null; // Allow the node to be freely moved afterward
        // d.fy = null; // Allow the node to be freely moved afterward
    }
}
