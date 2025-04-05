#!/usr/bin/env node
"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const taskSorter_1 = require("./taskSorter");
function main() {
    const args = process.argv.slice(2);
    if (args.length === 0 || args[0] === '--help') {
        console.log('Usage: asana-tasks-sorter [options]');
        console.log('Options:');
        console.log('  --help    Show this help message');
        return;
    }
    try {
        (0, taskSorter_1.sortTasks)(args);
    }
    catch (error) {
        console.error('Error:', error instanceof Error ? error.message : String(error));
        process.exit(1);
    }
}
main();
