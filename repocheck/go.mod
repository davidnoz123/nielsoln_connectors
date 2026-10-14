// repocheck -- the checks that can only be made from outside a single module.
//
// Its own module because a repo-wide check should not be owned by one
// connector: putting it in `shimp` or `chrome_sessions` would make an
// arbitrary folder responsible for its siblings. A program is also auditable
// on the same terms as everything else here, which a loose script is not.
//
// ⚠️ A PROGRAM, so it requires nothing, which is what s55's maxim 1 asks of
// anything that is not a leaf. Its `require` block is empty and stays empty.
module github.com/davidnoz123/nielsoln_connectors/repocheck

go 1.27
