async function callDemoApi(path) {
  const response = await fetch(path);
  document.querySelector("#result").textContent = await response.text();
}
