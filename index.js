// Set the theme
const storedTheme = window.matchMedia("(prefers-color-scheme: dark)").matches ? "luxury" : "bumblebee";
document.documentElement.setAttribute('data-theme', storedTheme);

function appendLocation(tParent, location) {
  const span = document.createElement("span");
  span.textContent = location
  span.className = "location"
  tParent.appendChild(span)
}

fetch('/locations.json')
  .then((response) => {
    if (!response.ok) {
      throw new Error('Network response was not "ok"');
    }
    return response.json();
  })
  .then((data) => {
    data.locations.forEach((location) => appendLocation(locationsEl, location.title))
  })
  .catch((error) => {
    console.error('There was a problem with the fetch operation:', error);
  });

const locationsEl = document.getElementById('locations');


// Collapsible Elements
const coll = document.getElementsByClassName("collapsible");
for (let i = 0; i < coll.length; i++) {
  coll[i].addEventListener("click", function() {
    this.classList.toggle("active");
    var content = this.nextElementSibling;
    if (content.style.display === "grid") {
      this.textContent = "Possible Locations +"
      content.style.display = "none";
    } else {
      this.textContent = "Possible Locations -"
      content.style.display = "grid";
    }
  });
}
