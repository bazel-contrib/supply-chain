const fs = require("node:fs");

const [apiInput, mainInput, output] = process.argv.slice(2);
if (!apiInput || !mainInput || !output) {
  throw new Error("expected two input files and an output path");
}

const bundle = [
  "(function () {",
  fs.readFileSync(apiInput, "utf8"),
  fs.readFileSync(mainInput, "utf8"),
  "})();",
  "",
].join("\n");
fs.writeFileSync(output, bundle);
