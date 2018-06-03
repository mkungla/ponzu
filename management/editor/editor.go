// Package editor enables users to create edit views from their content
// structs so that admins can manage content
package editor

import (
	"bytes"
	"log"
	"net/http"
)

var Role string

// Editable ensures data is editable
type Editable interface {
	MarshalEditor() ([]byte, error)
}

// Mergeable allows external post content to be approved and published through
// the public-facing API
type Mergeable interface {
	// Approve copies an external post to the internal collection and triggers
	// a re-sort of its content type posts
	Approve(http.ResponseWriter, *http.Request) error
}

// Editor is a view containing fields to manage content
type Editor struct {
	ViewBuf *bytes.Buffer
}

// Field is used to create the editable view for a field
// within a particular content struct
type Field struct {
	View []byte
	Role []string
}

func hiddenFor(roles []string) bool {
	if Role == "admin" {
		return false
	}
	for _, role := range roles {
		if role == Role {
			return true
		}
	}
	return false
}

// Form takes editable content and any number of Field funcs to describe the edit
// page for any content struct added by a user
func Form(post Editable, fields ...Field) ([]byte, error) {
	editor := &Editor{}

	editor.ViewBuf = &bytes.Buffer{}
	_, err := editor.ViewBuf.WriteString(`<table><tbody class="row"><tr class="col s8 editor-fields"><td class="col s12">`)
	if err != nil {
		log.Println("Error writing HTML string to editor Form buffer")
		return nil, err
	}

	for _, f := range fields {
		addFieldToEditorView(editor, f)

	}

	_, err = editor.ViewBuf.WriteString(`</td></tr>`)
	if err != nil {
		log.Println("Error writing HTML string to editor Form buffer")
		return nil, err
	}

	// content items with Item embedded have some default fields we need to render
	_, err = editor.ViewBuf.WriteString(`<tr class="col s4 default-fields"><td class="col s12">`)
	if err != nil {
		log.Println("Error writing HTML string to editor Form buffer")
		return nil, err
	}

	publishTime := `
<div class="row content-only __ponzu">
	<div class="input-field col s6">
		<label class="active">KK</label>
		<select class="month __ponzu browser-default">
			<option value="1">Jaanuar</option>
			<option value="2">Veebruar</option>
			<option value="3">Märts</option>
			<option value="4">Aprill</option>
			<option value="5">Mai</option>
			<option value="6">Juuni</option>
			<option value="7">Juuli</option>
			<option value="8">August</option>
			<option value="9">September</option>
			<option value="10">Oktoober</option>
			<option value="11">November</option>
			<option value="12">Detsember</option>
		</select>
	</div>
	<div class="input-field col s2">
		<label class="active">PP</label>
		<input value="" class="day __ponzu" maxlength="2" type="text" placeholder="PP" />
	</div>
	<div class="input-field col s4">
		<label class="active">AAAA</label>
		<input value="" class="year __ponzu" maxlength="4" type="text" placeholder="AAAA" />
	</div>
</div>

<div class="row content-only __ponzu">
	<div class="input-field col s3">
		<label class="active">TT</label>
		<input value="" class="hour __ponzu" maxlength="2" type="text" placeholder="TT" />
	</div>
	<div class="col s1">:</div>
	<div class="input-field col s3">
		<label class="active">MM</label>
		<input value="" class="minute __ponzu" maxlength="2" type="text" placeholder="MM" />
	</div>
	<div class="input-field col s4">
		<label class="active">Period</label>
		<select class="period __ponzu browser-default">
			<option value="AM">AM</option>
			<option value="PM">PM</option>
		</select>
	</div>
</div>
	`

	_, err = editor.ViewBuf.WriteString(publishTime)
	if err != nil {
		log.Println("Error writing HTML string to editor Form buffer")
		return nil, err
	}

	err = addPostDefaultFieldsToEditorView(post, editor)
	if err != nil {
		return nil, err
	}

	var submit string
	if Role == "admin" {
		submit = `
	<div class="input-field post-controls">
		<button class="right waves-effect waves-light btn green save-post" type="submit">Salvesta</button>
		<button class="right waves-effect waves-light btn red delete-post" type="submit">Kustuta</button>
	</div>
	`
	} else {
		submit = `
	<div class="input-field post-controls">
		<button class="right waves-effect waves-light btn green save-post" type="submit">Salvesta</button>
	</div>
	`
	}

	_, ok := post.(Mergeable)
	if ok {
		submit +=
			`
<div class="row external post-controls">
	<div class="col s12 input-field">
		<button class="right waves-effect waves-light btn blue approve-post" type="submit">Kinnitage</button>
		<button class="right waves-effect waves-light btn grey darken-2 reject-post" type="submit">Keeldu</button>
	</div>
	<label class="approve-details right-align col s12">See sisu on heakskiidu ootel. Klikkides "Kinnita" avaldatakse see kohe. Klõpsates "Keeldu", kustutatakse see.</label>
</div>
`
	}

	script := `
<script>
	$(function() {
		var form = $('form'),
			save = form.find('button.save-post'),
			del = form.find('button.delete-post'),
			external = form.find('.post-controls.external'),
			id = form.find('input[name=id]'),
			timestamp = $('.__ponzu.content-only'),
			slug = $('input[name=slug]');

		// hide if this is a new post, or a non-post editor page
		if (id.val() === '-1' || form.attr('action') !== '/admin/edit') {
			del.hide();
			external.hide();
		}

		// hide approval if not on a pending content item
		if (getParam('status') !== 'pending') {
			external.hide();
		}

		// no timestamp, slug visible on addons
		if (form.attr('action') === '/admin/addon') {
			timestamp.hide();
			slug.parent().hide();
		}

		save.on('click', function(e) {
			e.preventDefault();

			if (getParam('status') === 'pending') {
				var action = form.attr('action');
				form.attr('action', action + '?status=pending')
			}

			form.submit();
		});

		del.on('click', function(e) {
			e.preventDefault();
			var action = form.attr('action');
			action = action + '/delete';
			form.attr('action', action);

			if (confirm("Palun kinnita:\n\nKas soovite kindlasti selle sisu kustutada?\nSeda ei saa taastada.")) {
				form.submit();
			}
		});

		external.find('button.approve-post').on('click', function(e) {
			e.preventDefault();
			var action = form.attr('action');
			action = action + '/approve';
			form.attr('action', action);

			form.submit();
		});

		external.find('button.reject-post').on('click', function(e) {
			e.preventDefault();
			var action = form.attr('action');
			action = action + '/delete?reject=true';
			form.attr('action', action);

			if (confirm("[Palun kinnita:\n\nKas olete kindel, et soovite selle postituse tagasi lükata?\nSee kustutab selle sisu ja seda ei saa taastada.")) {
				form.submit();
			}
		});
	});
</script>
`
	_, err = editor.ViewBuf.WriteString(submit + script + `</td></tr></tbody></table>`)
	if err != nil {
		log.Println("Error writing HTML string to editor Form buffer")
		return nil, err
	}

	return editor.ViewBuf.Bytes(), nil
}

func addFieldToEditorView(e *Editor, f Field) (err error) {
	if hiddenFor(f.Role) {
		f.View = append([]byte(`<div style="display:none;">`), f.View...)
		f.View = append(f.View, []byte("</div>")...)
	}
	_, err = e.ViewBuf.Write(f.View)
	if err != nil {
		log.Println("Error writing field view to editor view buffer")
		return err
	}

	return nil
}

func addPostDefaultFieldsToEditorView(p Editable, e *Editor) error {
	defaults := []Field{
		{
			View: Input("Slug", p, map[string]string{
				"label":       "URL Slug",
				"type":        "text",
				"disabled":    "true",
				"placeholder": "Määratakse automaatselt",
			}),
		},
		{
			View: Timestamp("Timestamp", p, map[string]string{
				"type":  "hidden",
				"class": "timestamp __ponzu",
			}),
		},
		{
			View: Timestamp("Updated", p, map[string]string{
				"type":  "hidden",
				"class": "updated __ponzu",
			}),
		},
	}

	for _, f := range defaults {
		err := addFieldToEditorView(e, f)
		if err != nil {
			return err
		}
	}

	return nil
}
