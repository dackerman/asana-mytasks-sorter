#!/usr/bin/env node

import { sortTasks } from './taskSorter';

function main() {
  const args = process.argv.slice(2);
  
  if (args.length === 0 || args[0] === '--help') {
    console.log('Usage: asana-tasks-sorter [options]');
    console.log('Options:');
    console.log('  --help    Show this help message');
    return;
  }

  try {
    sortTasks(args);
  } catch (error) {
    console.error('Error:', error instanceof Error ? error.message : String(error));
    process.exit(1);
  }
}

main();