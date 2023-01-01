const form = document.getElementById("contracts-form");
const spinner = document.getElementById("loading-spinner");
document.getElementById("contract-type-button").addEventListener("click", function() {
    const address = document.getElementById("contract-address-input").value;
    fetch(`/checkContractStandard/${address}`)
        .then(response => response.json())
        .then(data => {
            document.getElementById("contract-type-result").innerText = data.standard;
            if (data.standard === "UNDEFINED") {
                document.getElementById("contract-type-result").style.color = "red";
            } else {
                document.getElementById("contract-type-result").style.color = "green";
            }
            document.getElementById("contract-type-container").style.display = "block";
        });
});
document.getElementById("verification-status-button").addEventListener("click", function() {
    const address = document.getElementById("contract-address-input").value;
    fetch(`/checkVerificationStatus/${address}`)
        .then(response => response.json())
        .then(data => {
            document.getElementById("verification-status-result").innerText = data.verified ? "Verified" : "Not verified";
            if (data.verified) {
            document.getElementById("verification-status-result").style.color = "green";
            } else {
            document.getElementById("verification-status-result").style.color = "red";
            }
            document.getElementById("verification-status-container").style.display = "block";
        });
});
document.getElementById("contracts-form").addEventListener("submit", function(event) {
event.preventDefault();
spinner.style.display = "block";

const startBlock = document.getElementById("start-block-input").value;
const endBlock = document.getElementById("end-block-input").value;
fetch(`/getContracts/${startBlock}/${endBlock}`)
    .then(response => response.json())
    .then(data => {
        // Get the list of contracts
        const contractTable = document.getElementById("contracts-table");

        // Clear the list of contracts
        contractTable.innerHTML = "";
        // Add the table header row
        contractTable.innerHTML = `
            <tr>
                <th>Address</th>
                <th>Standard</th>
                <th>Verified</th>
                <th>Transaction</th>
                <th>Block</th>
            </tr>
        `;
        // Iterate through the contracts
        // Add a row for each new contract
        data.forEach(contract => {
            const rowHTML = `
                <tr>
                    <td>${contract.address}</td>
                    <td>${contract.standard}</td>
                    <td>${contract.verified ? "Verified" : "Not verified"}</td>
                    <td>${contract.transaction}</td>
                    <td>${contract.block}</td>
                </tr>
            `;
            document.getElementById("contracts-table").insertAdjacentHTML("beforeend", rowHTML);
        });
        
        // Show the list of contracts
        document.getElementById("contracts-container").style.display = "block";
        spinner.style.display = "none";

    });
});
document.getElementById("contracts-form2").addEventListener("submit", function(event) {
event.preventDefault();
spinner.style.display = "block";
// Get the start and end times from the form inputs
const startTime = document.getElementById("start-time-input").value;
const endTime = document.getElementById("end-time-input").value;
// Send a request to the getContractsByTime endpoint with the start and end times as parameters
fetch(`/getContractsByTime/${startTime}/${endTime}`)
    .then(response => response.json())
    .then(data => {
    // Get the list of contracts
    const contractTable = document.getElementById("contracts-table");

    // Clear the list of contracts
    contractTable.innerHTML = "";		
        // Add the table header row
contractTable.innerHTML = `
    <tr>
    <th>Address</th>
    <th>Standard</th>
    <th>Verified</th>
    <th>Transaction</th>
    <th>Block</th>
    </tr>
`;
// Iterate through the contracts
// Add a row for each new contract
data.forEach(contract => {
    const rowHTML = `
    <tr>
        <td>${contract.address}</td>
        <td>${contract.standard}</td>
        <td>${contract.verified ? "Verified" : "Not verified"}</td>
        <td>${contract.transaction}</td>
        <td>${contract.block}</td>
    </tr>
    `;
    document.getElementById("contracts-table").insertAdjacentHTML("beforeend", rowHTML);
});
// Hide the loading spinner
spinner.style.display = "none";
// Show the contract table
document.getElementById("contracts-container").style.display = "block";
});
});