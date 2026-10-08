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

            const protectedResource =
                (
                    resource.type === "Pod" &&
                    resource.name.startsWith("platformpilot-")
                ) ||
                (
                    resource.type === "Service" &&
                    resource.name === "platformpilot"
                );

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
                            protectedResource
                                ? `<span class="protected">Protected</span>`
                                : `<button
                                    class="delete-button resource-delete"
                                    data-type="${resource.type}"
                                    data-name="${resource.name}">
                                    Delete
                                   </button>`
                        }
                    </td>
                </tr>
            `;
        }).join("");

        // Attach delete handlers after rendering
        document
            .querySelectorAll(".resource-delete")
            .forEach(button => {

                button.addEventListener("click", () => {

                    const type = button.dataset.type;
                    const name = button.dataset.name;

                    deleteResource(type, name);
                });
            });

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

async function deleteResource(type, name) {

    const confirmed = confirm(
        `Delete ${type} "${name}"?`
    );

    if (!confirmed) {
        return;
    }

    try {

        const response = await fetch(
            `/api/v1/resources/${type}/${encodeURIComponent(name)}`,
            {
                method: "DELETE"
            }
        );

        if (!response.ok) {
            throw new Error(await response.text());
        }

        await loadApplications();

    } catch (error) {

        alert(
            `Failed to delete ${type} ${name}: ${error.message}`
        );
    }
}
