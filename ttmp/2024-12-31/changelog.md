# HTML Help Templates

Added HTML templates using the templ templating language to provide HTML-formatted help output.

- Created HTML versions of all help templates in pkg/help/templates/html/
- Added proper HTML structure and semantic elements for better accessibility
- Maintained all functionality from the original templates while adding HTML formatting
- Added CSS classes for styling customization

# HTML Rendering Support

Added HTML rendering functionality to support the new HTML templates.

- Created html-render.go to handle HTML rendering using templ
- Added conversion functions between internal types and HTML template types
- Implemented parallel rendering paths for both Markdown and HTML output
- Maintained feature parity with existing Markdown rendering 

# Help System Type Reorganization

Reorganized help system types for better maintainability and type safety.

- Created central types package for all help data structures
- Organized types hierarchically based on their relationships
- Added builder functions for type construction and conversion
- Simplified template components by using shared type definitions
- Improved type safety and reduced code duplication 