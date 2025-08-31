document.addEventListener("DOMContentLoaded", function () {
  console.log("PMaaS Platform loaded");

  // --- Dashboard Animations & Charts ---

  // Soft entrance for cards
  const fadeIn = (els) => {
    els.forEach((el, i) => {
      el.style.opacity = "0";
      el.style.transform = "translateY(10px) scale(.98)";
      setTimeout(() => {
        el.style.transition =
          "opacity .55s cubic-bezier(.29,.63,.44,.75), transform .55s cubic-bezier(.29,.63,.44,.75)";
        el.style.opacity = "1";
        el.style.transform = "translateY(0) scale(1)";
      }, i * 90);
    });
  };

  const cards = document.querySelectorAll("sl-card");
  if (cards.length > 0) {
    fadeIn(Array.from(cards));
  }

  // Mini charts (Chart.js)
  const makeSpark = (id, data, color = "rgb(52,211,153)") => {
    const el = document.getElementById(id);
    if (!el) return; // Only run if the chart canvas exists
    new Chart(el.getContext("2d"), {
      type: "line",
      data: {
        labels: data.map((_, i) => i + 1),
        datasets: [
          {
            data,
            borderColor: color,
            borderWidth: 2,
            pointRadius: 0,
            tension: 0.35,
            fill: false,
          },
        ],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: { legend: { display: false }, tooltip: { enabled: false } },
        scales: { x: { display: false }, y: { display: false } },
        elements: { line: { capBezierPoints: true } },
      },
    });
  };

  // Fake sample sparkline data (replace with live later)
  makeSpark("sparkTotal", [3, 4, 4, 5, 4, 4, 4], "rgb(94,234,212)");
  makeSpark("sparkOcc", [1, 2, 2, 2, 2, 2, 2], "rgb(74,222,128)");
  makeSpark("sparkVac", [2, 1, 1, 1, 1, 1, 1], "rgb(248,113,113)");
  makeSpark(
    "sparkRev",
    [3800, 4200, 4800, 5000, 5300, 5300, 5300],
    "rgb(34,197,94)",
  );

  // --- Add Property Modal Logic ---

  const addPropertyBtn = document.getElementById("addPropertyBtn");
  const addPropertyModal = document.getElementById("addPropertyModal");
  const cancelBtn = document.getElementById("cancelBtn");
  const addPropertyForm = document.getElementById("addPropertyForm");

  // Show the modal when the "Add New Property" button is clicked
  if (addPropertyBtn) {
    addPropertyBtn.addEventListener("click", () => {
      if (addPropertyModal) {
        addPropertyModal.classList.remove("hidden");
      }
    });
  }

  // A function to hide the modal
  const hideModal = () => {
    if (addPropertyModal) {
      addPropertyModal.classList.add("hidden");
    }
  };

  // Hide the modal when the "Cancel" button is clicked
  if (cancelBtn) {
    cancelBtn.addEventListener("click", hideModal);
  }

  // Hide the modal when clicking on the background overlay
  if (addPropertyModal) {
    addPropertyModal.addEventListener("click", (event) => {
      if (event.target === addPropertyModal) {
        hideModal();
      }
    });
  }

  // Handle form submission
  if (addPropertyForm) {
    addPropertyForm.addEventListener("submit", (event) => {
      event.preventDefault(); // Prevent default form submission
      const formData = new FormData(addPropertyForm);
      const propertyData = Object.fromEntries(formData.entries()); // Extract form data
      console.log(
        "Attempting to add new property to the database:",
        propertyData,
      );
      // Send the property data to the server-side API endpoint
      // The server-side API (e.g., Node.js, Python, PHP) will then interact with the database
      fetch("/properties", {
        method: "POST", // Use POST for creating new resources
        headers: {
          "Content-Type": "application/json", // Indicate that the body is JSON
        },
        body: JSON.stringify(propertyData), // Convert JS object to JSON string
      })
        .then((response) => {
          if (response.ok) {
            // Property successfully added (HTTP status 2xx)
            console.log("Property added successfully to the database.");
            // Reload the page to show the newly added property in the list
            window.location.reload();
          } else {
            // Server responded with an error (e.g., 400, 500)
            // Attempt to parse error message from server response
            response
              .json()
              .then((errorBody) => {
                console.error(
                  "Failed to add property. Server status:",
                  response.status,
                  "Error details:",
                  errorBody,
                );
                alert(
                  `Failed to add property: ${errorBody.message || "An unexpected server error occurred."}`,
                );
              })
              .catch(() => {
                // Handle cases where the error response is not JSON
                console.error(
                  "Failed to add property. Server status:",
                  response.status,
                  "No specific error details from server.",
                );
                alert(
                  "Failed to add property. Please check console for details.",
                );
              });
          }
        })
        .catch((error) => {
          // Network error or problem with the fetch operation itself
          console.error("Network error or problem submitting form:", error);
          alert(
            "An error occurred while connecting to the server. Please check your internet connection.",
          );
        })
        .finally(() => {
          // Always reset the form and hide the modal after submission attempt
          addPropertyForm.reset();
          hideModal();
        });
    });
  }
});
