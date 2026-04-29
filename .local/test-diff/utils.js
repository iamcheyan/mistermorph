function greet(name) {
    console.log("Hello, " + name + "!");
}

function sum(arr) {
    let total = 0;
    for (let i = 0; i < arr.length; i++) {
        total += arr[i];
    }
    return total;
}

module.exports = { greet, sum };
