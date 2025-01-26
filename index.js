// Set the theme
const storedTheme = window.matchMedia("(prefers-color-scheme: dark)").matches ? "luxury" : "bumblebee";
document.documentElement.setAttribute('data-theme', storedTheme);

// Change the picture based on the system theme
const isDark = window.matchMedia("(prefers-color-scheme: dark)").matches;
const splash = document.getElementById("splash-container");
const img = document.createElement("img");

if (isDark) {
  img.src = "./public/crabspy-splash-night.png";
  splash.appendChild(img)
} else {
  img.src = "./public/crabspy-splash-day.png";
  splash.appendChild(img)
}



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
const rules = document.getElementById('rules');
const rulesText = document.getElementById('rules-text');
const locs = document.getElementById('possible-locations');

rules.addEventListener('click', () => {
  if (rulesText.style.display == "none") {
    rulesText.style.display = "block"
    locationsEl.style.display = "none"
  } else {
    rulesText.style.display = "none"
  }
})

locs.addEventListener('click', () => {
  if (locationsEl.style.display == "none") {
    locationsEl.style.display = "grid"
    rulesText.style.display = "none"
  } else {
    locationsEl.style.display = "none"
  }
})
