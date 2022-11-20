export function sortByProperty<T>(
  skillArray: Array<T>,
  getSortingProperty: (x: T, y: T) => [any, any]
): Array<T> {
    return skillArray.sort((a: T, b: T) => {
        const [aProperty, bProperty] = getSortingProperty(a, b);
        if (aProperty < bProperty) return -1;
        if (aProperty > bProperty) return 1;
        return 0;
    });
}