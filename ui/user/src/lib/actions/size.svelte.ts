export function transitionParentHeight(node: HTMLElement, fn: () => any) {
    const onsizechange = () => {

        // Debounce the resize calculation
        requestAnimationFrame(() => {
            if (!node.parentElement) {
                return
            }
            node.parentElement!.style.height = node.scrollHeight + 'px';
        })
    }

    onsizechange();

    const observer = new ResizeObserver(onsizechange);

    $effect(() => {
        // Recalculate the parent height programmatically
        fn();

        onsizechange();

        observer.observe(node)

        return () => {
            observer.disconnect()

            if (node.parentElement) {
                // Use the computed style to get the pixel value of the min-height; 
                const style = getComputedStyle(node.parentElement);

                node.parentElement!.style.height = style.minHeight;
            }
        }
    })

}