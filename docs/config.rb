###
# Page options, layouts, aliases and proxies
###

# Per-page layout changes:
#
# With no layout
page '/*.xml', layout: false
page '/*.json', layout: false
page '/*.txt', layout: false

# With alternative layout
# page "/path/to/file.html", layout: :otherlayout

# Proxy pages (http://middlemanapp.com/basics/dynamic-pages/)
# proxy "/this-page-has-no-template.html", "/template-file.html", locals: {
#  which_fake_page: "Rendering a fake page with a local variable" }

# General configuration
activate :asciidoc, attributes: {"icons" => "font"}

class AsciidoctorTweaks < Middleman::Extension
  def initialize app, options_hash = {}, &block
    super
    app.before do
      app.config[:asciidoc][:base_dir] = app.source_dir
      true
    end
  end
end

::Middleman::Extensions.register :asciidoctor_tweaks, AsciidoctorTweaks

activate :asciidoctor_tweaks

# Reload the browser automatically whenever files change
configure :development do
  activate :livereload
end

activate :syntax, :line_numbers => true

# Build-specific configuration
configure :build do
  # Minify CSS on build
  # activate :minify_css

  # Minify Javascript on build
  # activate :minify_javascript
end
