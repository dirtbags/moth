MOTH Metadata
============



Standard Metadata
-----------------

The following are considered "standard" MOTH metadata.
Clients *should* check for, 
and take appropriate action on,
all of these metadata names.


| name | description | permitted values | example |
| --- | --- | --- |  --- |
| author | Puzzle author(s). | free text [(ref)](https://developer.mozilla.org/en-US/docs/Web/HTML/Element/meta/name) | `Neale Pickett` |
| moth.style | Whether the client should inject a style sheet. Default: `inherit` | `override`, `inherit` | `override` |
| moth.answerhash | Answer hash, used for "possibly correct" check in client. | MOTHv5: first 8 characters of answer's SHA1 checksum | `a5b6bb92` |
| moth.answerpattern | Answer pattern, to use as `pattern` attribute of `<input>` element for answer. | Regular Expression [(ref)](https://developer.mozilla.org/en-US/docs/Web/HTML/Attributes/pattern) | `\w{3,16}`
| moth.ksa | [NICE KSA](https://niccs.cisa.gov/workforce-development/nice-framework) achieved by completing this puzzle. | NICE KSA identifier | `K0052` |
| moth.objective | Learning objective of this puzzle. | free text | `Count in octal` |
| moth.success.acceptable | The minimum work required to be considered successfully understanding this puzzle's concepts | free text | `Recognize pattern` |
| moth.success.mastery | The work required to be considered mastering this puzzle's concepts | free text | `Understand 8s place in octal` |


Standard Debugging Metadata
----------------

These metadata names are for debugging purposes.
The *must not* be present in a production instance.

| name | description | permitted values | example |
| --- | --- | --- | --- |
| moth.debug.answer | An accepted answer | free text | `pink hat horse race` |
| moth.debug.summary | A summary of the puzzle, to help staff remember what it is | free text | `Hidden white text in the rendered image` |
| moth.debug.hint | A hint that staff can provide to participants | free text | `This puzzle can be solved by a grade school student with no special tools` |
| moth.debug.notes | Notes to staff intended to help better understand the puzzle | free text | `We used this image because Scott likes tigers` |
| moth.debug.log | A log message | free text | `iterations: 5` |
| moth.debug.errors | Error messages | free text | `unable to open foo.bin` |


Client Metadata
-----------

Clients wishing to implement additional metadata
*should* either submit a merge request to this document,
or use a `moth/$client.` prefix.
For example, the "tofu" client might use a
`moth/tofu.difficulty` name.
