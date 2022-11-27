export function sortByProperty<T>(
  skillArray: Array<T>,
  getPropertyToSortBy: (element: T) => any
): Array<T> {
    return skillArray.sort((a: T, b: T) => {
        const aProperty = getPropertyToSortBy(a);
        const bProperty = getPropertyToSortBy(b);
        if (aProperty < bProperty) return -1;
        if (aProperty > bProperty) return 1;
        return 0;
    });
}