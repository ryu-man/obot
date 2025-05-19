export function clickOutside(element: HTMLElement, onClickOutside: (event: Event) => void) {
	function checkClickOutside(event: Event) {
		if (element?.contains(event.target as HTMLElement)) return;

		onClickOutside(event);
	}

	// <dialog> called with showModal()
	const isModalDialog =
		element.tagName.toLowerCase() === 'dialog' &&
		(element as HTMLDialogElement).open &&
		(element as HTMLDialogElement).showModal !== undefined;

	if (!isModalDialog) {
		document.addEventListener('click', checkClickOutside);
	} else {
		element.addEventListener('click', checkClickOutside);
	}

	return {
		destroy() {
			if (!isModalDialog) {
				document.removeEventListener('click', checkClickOutside);
			} else {
				element.removeEventListener('click', checkClickOutside);
			}
		}
	};
}
