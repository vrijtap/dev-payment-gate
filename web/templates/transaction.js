/**
 * Function to submit a payment
 */
function pay() {
    // Find the element with id 'fakePay-submit'
    var fakePaySubmitButton = document.getElementById("fakePay-submit");
    
    // Remove the event listener to disable further clicks
    fakePaySubmitButton.removeEventListener("click", pay);

    // Define the payment url
    let url = `/transaction/{{.ID}}`;

    // Perform the request
    fetch(url, {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        }
    })
    .then(response => response.json()) // Parse the JSON response
    .then(data => {
        // Check if the response contains a 'url' field
        if (data && data.url) {
            // Redirect the user to the received URL
            window.location.href = data.url;
        } else {
            console.error('Invalid response format');
        }
    })
    .catch(error => {
        console.error('Error:', error);
    })
    .finally(() => {
        // Add back the event listener to re-enable clicks after the operation is completed
        fakePaySubmitButton.addEventListener("click", pay);
    });
}

/**
 * Event to add all of the javascript functionality after loading the DOM
 */
document.addEventListener("DOMContentLoaded", function() {
    // Find the element with id 'fakePay-submit'
    var fakePaySubmitButton = document.getElementById("fakePay-submit");

    // Add an onclick event listener to pay
    fakePaySubmitButton.addEventListener("click", pay);

    // Show the fakepay element now that it has functionality
    document.querySelector(".fakePay").style.display = "block";
});
