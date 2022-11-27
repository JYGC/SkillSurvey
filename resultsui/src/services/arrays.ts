export function sortByProperty<T>(
  skillArray: Array<T>,
  getSortingProperty: (element: T) => any
): Array<T> {
    return skillArray.sort((a: T, b: T) => {
        const aProperty = getSortingProperty(a);
        const bProperty = getSortingProperty(b);
        if (aProperty < bProperty) return -1;
        if (aProperty > bProperty) return 1;
        return 0;
    });
}