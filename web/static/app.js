async function loadApplications() {
    const table = document.getElementById("applicationsTable");

    table.innerHTML = `
        <tr>
            <td colspan="5" class="empty">Loading resources...</td>
        </tr>
    `;

    try {
        const response = await fetch("/api/v1/resources");

        if (!response.ok) {
            throw new Error("Failed to load Kubernetes resources");
        }

        const data = await response.json();

        const resources = [
            ...(data.deployments || []),
            ...(data.pods || []),
            ...(data.services || [])
        ];

        document.getElementById("applicationCount").textContent =
            data.deployments?.length || 0;

        document.getElementById("runningCount").textContent =
            (data.pods || []).filter(
                pod => pod.status === "Running"
            ).length;

        if (resources.length === 0) {
            table.innerHTML = `
                <tr>
                    <td colspan="5" class="empty">
                        No resources found.
                    </td>
                </tr>
            `;
            return;
        }

        table.innerHTML = resources.map(resource => {

            let replicaText = "-";

            if (resource.type === "Deployment") {
                replicaText =
                    `${resource.available}/${resource.replicas}`;
            }

            return `
                <tr>
                    <td>
                        <strong>${resource.name}</strong>
                    </td>

                    <td>${resource.type}</td>

                    <td>
                        <span class="status-badge">
                            ${resource.status}
                        </span>
                    </td>

                    <td>${replicaText}</td>

                    <td>
    			${
        			(
           		 resource.type === "Pod" &&
           		 resource.name.startsWith("platformpilot-")
       			 ) ||
       			 (
           		 resource.type === "Service" &&
           		 resource.name === "platformpilot"
       			 )
           		 ? "-"
           		 : `<button
               		 class="delete-button"
               		 onclick="deleteResource('${resource.type}', '${resource.name}')">
              		 Delete
           		 </button>`
   			 }
		  </td>
                </tr>
            `;
        }).join("");

    } catch (error) {

        table.innerHTML = `
            <tr>
                <td colspan="5" class="empty">
                    ${error.message}
                </td>
            </tr>
        `;
    }
}


async function deleteApplication(name) {

    const confirmed = confirm(
        `Delete application "${name}"?`
    );

    if (!confirmed) {
        return;
    }

    try {

        const response = await fetch(
            `/api/v1/applications/${name}`,
            {
                method: "DELETE"
            }
        );

        if (!response.ok) {
            throw new Error(
                await response.text()
            );
        }

        await loadApplications();

    } catch (error) {

        alert(
            `Failed to delete ${name}: ${error.message}`
        );
    }
}


function openDeployModal() {
    document
        .getElementById("deployModal")
        .classList.remove("hidden");
}


function closeDeployModal() {
    document
        .getElementById("deployModal")
        .classList.add("hidden");
}


document
    .getElementById("resourceType")
    .addEventListener("change", function () {

        const recommendation =
            document.getElementById("recommendation");

        if (this.value === "pod") {

            recommendation.innerHTML = `
                <strong>⚠️ Recommendation: Deployment</strong>

                <p>
                    A standalone Pod is usually not recommended
                    for long-running applications.
                </p>

                <ul>
                    <li>Pods are not automatically recreated.</li>
                    <li>No replica management.</li>
                    <li>No rolling updates.</li>
                </ul>

                <p>
                    Consider using a Deployment instead.
                </p>
            `;

        } else if (this.value === "service") {

            recommendation.innerHTML = `
                <strong>💡 Service</strong>

                <p>
                    A Service provides a stable network endpoint
                    for your Kubernetes application.
                </p>
            `;

        } else {

            recommendation.innerHTML = `
                <strong>💡 Recommended</strong>

                <p>
                    Deployment is recommended for most
                    long-running applications.
                </p>

                <ul>
                    <li>Self-healing</li>
                    <li>Replica management</li>
                    <li>Rolling updates</li>
                    <li>Easy scaling</li>
                </ul>
            `;
        }
    });


document
    .getElementById("deployForm")
    .addEventListener("submit", async function (event) {

        event.preventDefault();

        const errorBox =
            document.getElementById("deployError");

        errorBox.classList.add("hidden");

        const payload = {
            name: document.getElementById("appName").value,
            image: document.getElementById("image").value,
            replicas: Number(
                document.getElementById("replicas").value
            )
        };

        try {

            const response = await fetch(
                "/api/v1/applications",
                {
                    method: "POST",

                    headers: {
                        "Content-Type": "application/json"
                    },

                    body: JSON.stringify(payload)
                }
            );

            if (!response.ok) {
                throw new Error(
                    await response.text()
                );
            }

            closeDeployModal();

            document
                .getElementById("deployForm")
                .reset();

            document.getElementById("replicas").value = 2;
            document.getElementById("image").value = "nginx:1.27";

            await loadApplications();

        } catch (error) {

            errorBox.textContent = error.message;

            errorBox.classList.remove("hidden");
        }
    });


loadApplications();

