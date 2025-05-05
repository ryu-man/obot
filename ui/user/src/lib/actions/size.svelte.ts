import gsap from "gsap";

export type AnimateHeightActionProps = {
    isActive?: () => boolean,
    vars?: Pick<gsap.TweenVars, 'duration' | 'delay' | 'ease'>
}

export function animateHeight(node: HTMLElement, { isActive, vars }: AnimateHeightActionProps) {
    let currentTween: gsap.core.Tween | undefined = undefined;

    $effect(() => {

        const { ease = undefined, delay = 0, duration = .2 } = vars ?? {};

        const onsizechange = () => {
            // Debounce the resize calculation
            currentTween?.pause?.();
            if (!node.parentElement) {
                return
            }




            if (isActive && isActive()) {
                currentTween = gsap.to(node.parentElement, { height: node.scrollHeight, duration, delay, ease })

            } else {
                gsap.to(node.parentElement, { height: 'auto', duration, delay, ease });
            }
        }

        onsizechange();

        const observer = new ResizeObserver(onsizechange);


        observer.observe(node)

        return () => {
            currentTween?.pause?.();
            observer.disconnect()

            gsap.to(node.parentElement, { height: 'auto', duration, delay, ease });
        }
    })

}