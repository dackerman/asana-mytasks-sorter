import { sortTasks } from '../src/taskSorter';

describe('sortTasks', () => {
  test('should sort task arguments alphabetically', () => {
    const args = ['c', 'a', 'b'];
    const result = sortTasks(args);
    expect(result).toEqual(['a', 'b', 'c']);
  });

  test('should return empty array when no tasks provided', () => {
    const args: string[] = [];
    const result = sortTasks(args);
    expect(result).toEqual([]);
  });
});